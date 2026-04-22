package storage

import (
	"cloud_disk/core/internal/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	openapiutil "github.com/alibabacloud-go/darabonba-openapi/v2/utils"
	stsclient "github.com/alibabacloud-go/sts-20150401/v2/client"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// OSSService 是当前项目中所有 OSS / STS 基础能力的统一入口。
//
// 这一层只负责和阿里云能力打交道，不直接关心业务表之间如何写入。
// 业务层通过它来完成：
// 1. 生成 objectKey
// 2. 为前端上传签发 STS 临时凭证
// 3. 生成预览所需的签名 URL
// 4. 读取文本预览片段
// 5. 删除重复对象、兼容旧 path 数据
type OSSService struct {
	cfg config.OSSConfig
}

// STSCredential 是发给前端的临时凭证结构。
// 前端只需要这一组信息就能直传 OSS，不需要接触后端长期 AK/SK。
type STSCredential struct {
	AccessKeyID     string `json:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"`
	SecurityToken   string `json:"security_token"`
	Expiration      string `json:"expiration"`
}

const (
	// 这里复用 OSS SDK 默认的环境变量名称，方便直接兼容本地已有配置。
	envAlibabaCloudAccessKeyID     = "OSS_ACCESS_KEY_ID"
	envAlibabaCloudAccessKeySecret = "OSS_ACCESS_KEY_SECRET"
	envAlibabaCloudSecurityToken   = "OSS_SESSION_TOKEN"

	// RoleArn / ExternalID 不属于 OSS SDK 凭证 provider 负责的内容，
	// 因此这里单独保留环境变量兜底。
	envAlibabaCloudRoleArn    = "ALIBABA_CLOUD_ROLE_ARN"
	envAlibabaCloudExternalID = "ALIBABA_CLOUD_EXTERNAL_ID"
)

// NewOSSService 在服务启动时统一补齐默认配置，减少业务代码的分支判断。
func NewOSSService(cfg config.OSSConfig) *OSSService {
	if cfg.UploadBaseDir == "" {
		cfg.UploadBaseDir = "cloud-disk"
	}
	if cfg.StsDurationSeconds == 0 {
		cfg.StsDurationSeconds = 3600
	}
	if cfg.PreviewExpireSeconds == 0 {
		cfg.PreviewExpireSeconds = 1800
	}
	return &OSSService{cfg: cfg}
}

// IssueUploadSTS 为某个上传会话签发“只针对一个 objectKey”的临时凭证。
//
// 这一步的核心目的是：
// 1. 前端只拿到短期可用的 STS，不拿后端永久 AK/SK
// 2. 前端上传范围被限制到当前 objectKey，不能随意写 OSS 其它路径
// 3. 一次上传绑定一个 upload_session，后续暂停/继续/续签都围绕这个会话进行
func (s *OSSService) IssueUploadSTS(ctx context.Context, sessionIdentity string, objectKey string) (*STSCredential, error) {
	if err := s.validateSTSConfig(); err != nil {
		return nil, err
	}

	// 这里使用后端长期凭证去请求 STS，然后把返回的临时凭证下发给前端。
	openapiCfg := &openapiutil.Config{
		AccessKeyId:     stringPtr(s.accessKeyID()),
		AccessKeySecret: stringPtr(s.accessKeySecret()),
		RegionId:        stringPtr(s.normalizedRegion()),
	}
	if token := s.securityToken(); token != "" {
		openapiCfg.SecurityToken = stringPtr(token)
	}

	client, err := stsclient.NewClient(openapiCfg)
	if err != nil {
		return nil, err
	}

	// 这里构造的是“会话级别”的最小权限策略：
	// 即使 RoleArn 本身更大，这次返回给前端的临时凭证也只允许操作当前 objectKey。
	policy, err := buildUploadPolicy(s.cfg.Bucket, objectKey)
	if err != nil {
		return nil, err
	}

	duration := s.cfg.StsDurationSeconds
	if duration < 900 {
		// STS 对最小有效期有要求，过短的值会被接口拒绝。
		duration = 900
	}

	// RoleSessionName 主要用于审计和排查问题时定位是哪次上传。
	req := (&stsclient.AssumeRoleRequest{}).
		SetRoleArn(s.roleArn()).
		SetRoleSessionName(buildRoleSessionName(sessionIdentity)).
		SetPolicy(policy).
		SetDurationSeconds(duration)

	if externalID := s.externalID(); externalID != "" {
		req.SetExternalId(externalID)
	}

	resp, err := client.AssumeRole(req)
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Body == nil || resp.Body.Credentials == nil {
		return nil, errors.New("STS credentials are empty")
	}

	return &STSCredential{
		AccessKeyID:     stringValue(resp.Body.Credentials.AccessKeyId),
		AccessKeySecret: stringValue(resp.Body.Credentials.AccessKeySecret),
		SecurityToken:   stringValue(resp.Body.Credentials.SecurityToken),
		Expiration:      stringValue(resp.Body.Credentials.Expiration),
	}, nil
}

// SignGetObjectURL 为预览类请求生成一个短时有效的 OSS 签名地址。
// 浏览器拿到这个地址后可以直接访问对象，而不需要把整个文件流量都打回业务服务。
func (s *OSSService) SignGetObjectURL(ctx context.Context, objectKey string, expires time.Duration, inlineName string) (string, error) {
	client, err := s.newPermanentClient()
	if err != nil {
		return "", err
	}

	req := &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.cfg.Bucket),
		Key:    oss.Ptr(objectKey),
	}
	if inlineName != "" {
		// 指定 inline disposition，浏览器更容易直接按预览方式处理。
		disposition := fmt.Sprintf("inline; filename*=UTF-8''%s", url.QueryEscape(inlineName))
		req.ResponseContentDisposition = oss.Ptr(disposition)
	}

	result, err := client.Presign(ctx, req, oss.PresignExpires(expires))
	if err != nil {
		return "", err
	}
	return result.URL, nil
}

// ReadObjectRange 只读取对象前面一小段内容，用于文本预览。
// 这样可以避免大文本文件一次性被完整加载。
func (s *OSSService) ReadObjectRange(ctx context.Context, objectKey string, maxBytes int64) ([]byte, bool, error) {
	client, err := s.newPermanentClient()
	if err != nil {
		return nil, false, err
	}

	// 先读取对象头部，后面可以判断当前返回内容是否被截断。
	head, err := client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(s.cfg.Bucket),
		Key:    oss.Ptr(objectKey),
	})
	if err != nil {
		return nil, false, err
	}

	req := &oss.GetObjectRequest{
		Bucket: oss.Ptr(s.cfg.Bucket),
		Key:    oss.Ptr(objectKey),
		Range:  oss.Ptr(fmt.Sprintf("bytes=0-%d", maxBytes-1)),
	}
	result, err := client.GetObject(ctx, req)
	if err != nil {
		return nil, false, err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, false, err
	}

	return body, head.ContentLength > int64(len(body)), nil
}

// HeadObject 主要用于上传完成时校验 OSS 上对象大小是否与会话一致。
func (s *OSSService) HeadObject(ctx context.Context, objectKey string) (*oss.HeadObjectResult, error) {
	client, err := s.newPermanentClient()
	if err != nil {
		return nil, err
	}

	return client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(s.cfg.Bucket),
		Key:    oss.Ptr(objectKey),
	})
}

// DeleteObject 主要用于并发上传场景下清理“重复产生的多余对象”。
func (s *OSSService) DeleteObject(ctx context.Context, objectKey string) error {
	client, err := s.newPermanentClient()
	if err != nil {
		return err
	}

	_, err = client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(s.cfg.Bucket),
		Key:    oss.Ptr(objectKey),
	})
	return err
}

// BuildObjectKey 统一生成 OSS 中的对象路径。
//
// 前端永远不能自己决定 objectKey，必须由后端生成并通过 STS 限制写入目标。
// 路径结构采用：上传目录 / 用户 identity / 日期 / 时间戳-文件名。
func (s *OSSService) BuildObjectKey(userIdentity string, fileName string) string {
	cleanName := sanitizeFileName(fileName)
	ext := path.Ext(cleanName)
	base := strings.TrimSuffix(cleanName, ext)
	if base == "" {
		base = "file"
	}
	timestamp := time.Now().Format("20060102")
	return path.Join(
		s.cfg.UploadBaseDir,
		userIdentity,
		timestamp,
		fmt.Sprintf("%d-%s%s", time.Now().UnixNano(), base, ext),
	)
}

// BuildObjectURL 把 objectKey 还原成标准 OSS 访问地址。
// 预览签名失败时，可作为兼容旧数据或问题排查的兜底结果。
func (s *OSSService) BuildObjectURL(objectKey string) string {
	endpoint := strings.TrimSpace(s.cfg.Endpoint)
	if endpoint == "" {
		endpoint = s.defaultEndpoint()
	}
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", s.cfg.Bucket, endpoint, objectKey)
}

// BrowserRegion 返回前端 ali-oss 需要的 region 格式。
// 前端 SDK 更习惯 oss-cn-hangzhou 这种完整写法。
func (s *OSSService) BrowserRegion() string {
	region := s.normalizedRegion()
	if strings.HasPrefix(region, "oss-") {
		return region
	}
	return "oss-" + region
}

// Endpoint 优先返回配置中的 endpoint，未配置时按 region 推导默认值。
func (s *OSSService) Endpoint() string {
	if strings.TrimSpace(s.cfg.Endpoint) != "" {
		return s.cfg.Endpoint
	}
	return s.defaultEndpoint()
}

func (s *OSSService) Bucket() string {
	return s.cfg.Bucket
}

// PreviewExpires 统一管理预览签名的有效期。
func (s *OSSService) PreviewExpires() time.Duration {
	return time.Duration(s.cfg.PreviewExpireSeconds) * time.Second
}

// GuessObjectKey 用于兼容历史 path 数据。
//
// 老数据里 repository_pool.path 可能是：
// 1. 纯 objectKey
// 2. bucket/objectKey
// 3. 完整 URL
//
// 这里尝试从前两种恢复 objectKey；如果是完整 URL，则交给上层继续走旧地址。
func (s *OSSService) GuessObjectKey(pathValue string) string {
	pathValue = strings.TrimSpace(pathValue)
	if pathValue == "" {
		return ""
	}
	if strings.HasPrefix(pathValue, "http://") || strings.HasPrefix(pathValue, "https://") {
		return ""
	}
	prefix := s.cfg.Bucket + "/"
	if strings.HasPrefix(pathValue, prefix) {
		return strings.TrimPrefix(pathValue, prefix)
	}
	return pathValue
}

// newPermanentClient 创建服务端长期使用的 OSS 客户端。
// 它主要被后端拿来做签名、预览读取、上传校验、对象清理等动作。
func (s *OSSService) newPermanentClient() (*oss.Client, error) {
	if err := s.validatePermanentConfig(); err != nil {
		return nil, err
	}

	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewStaticCredentialsProvider(s.accessKeyID(), s.accessKeySecret(), s.securityToken())).
		WithRegion(s.normalizedRegion()).
		WithEndpoint(s.Endpoint())

	return oss.NewClient(cfg), nil
}

// validatePermanentConfig 校验 OSS 基础配置是否可用。
func (s *OSSService) validatePermanentConfig() error {
	if s.cfg.Bucket == "" {
		return errors.New("OSS bucket is not configured")
	}
	if s.normalizedRegion() == "" {
		return errors.New("OSS region is not configured")
	}
	if s.accessKeyID() == "" || s.accessKeySecret() == "" {
		return errors.New("OSS access key is not configured")
	}
	return nil
}

// validateSTSConfig 在永久凭证校验基础上额外检查 RoleArn。
// 因为 AssumeRole 必须明确“当前要扮演哪个 RAM 角色”。
func (s *OSSService) validateSTSConfig() error {
	if err := s.validatePermanentConfig(); err != nil {
		return err
	}
	if s.roleArn() == "" {
		return errors.New("OSS role ARN is not configured")
	}
	return nil
}

// normalizedRegion 去掉可能存在的 oss- 前缀，统一内部处理格式。
func (s *OSSService) normalizedRegion() string {
	return strings.TrimPrefix(strings.TrimSpace(s.cfg.Region), "oss-")
}

// defaultEndpoint 在未配置自定义 endpoint 时，按 region 拼接标准 OSS 域名。
func (s *OSSService) defaultEndpoint() string {
	region := s.normalizedRegion()
	if region == "" {
		return ""
	}
	return fmt.Sprintf("oss-%s.aliyuncs.com", region)
}

// 下面这些读取函数统一遵循“配置优先，环境变量兜底”的规则。
// 这样部署时可写配置，本地调试时也能继续复用环境变量。
func (s *OSSService) accessKeyID() string {
	if value := strings.TrimSpace(s.cfg.AccessKeyId); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv(envAlibabaCloudAccessKeyID))
}

func (s *OSSService) accessKeySecret() string {
	if value := strings.TrimSpace(s.cfg.AccessKeySecret); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv(envAlibabaCloudAccessKeySecret))
}

func (s *OSSService) securityToken() string {
	return strings.TrimSpace(os.Getenv(envAlibabaCloudSecurityToken))
}

func (s *OSSService) roleArn() string {
	if value := strings.TrimSpace(s.cfg.RoleArn); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv(envAlibabaCloudRoleArn))
}

func (s *OSSService) externalID() string {
	if value := strings.TrimSpace(s.cfg.ExternalID); value != "" {
		return value
	}
	return strings.TrimSpace(os.Getenv(envAlibabaCloudExternalID))
}

// buildUploadPolicy 构造“单次上传会话”的最小权限策略。
func buildUploadPolicy(bucket string, objectKey string) (string, error) {
	resource := fmt.Sprintf("acs:oss:*:*:%s/%s", bucket, objectKey)
	policy := map[string]any{
		"Version": "1",
		"Statement": []map[string]any{
			{
				"Effect": "Allow",
				"Action": []string{
					"oss:PutObject",
					"oss:InitiateMultipartUpload",
					"oss:UploadPart",
					"oss:CompleteMultipartUpload",
					"oss:AbortMultipartUpload",
					"oss:ListParts",
				},
				"Resource": []string{resource},
			},
		},
	}
	raw, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// buildRoleSessionName 生成 STS 角色会话名。
func buildRoleSessionName(sessionIdentity string) string {
	name := "upload-" + strings.ReplaceAll(sessionIdentity, "-", "")
	if len(name) > 64 {
		return name[:64]
	}
	return name
}

// sanitizeFileName 对文件名做基础清洗，避免把路径分隔符带进 objectKey。
func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")
	if name == "" {
		return "file"
	}
	return name
}

func stringPtr(v string) *string {
	return &v
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
