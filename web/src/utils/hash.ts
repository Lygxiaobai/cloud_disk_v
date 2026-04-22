import SparkMD5 from "spark-md5";

export async function computeFileMD5(
  file: File,
  onProgress?: (progress: number) => void,
  chunkSize = 2 * 1024 * 1024,
): Promise<string> {
  const spark = new SparkMD5.ArrayBuffer();
  const totalChunks = Math.max(1, Math.ceil(file.size / chunkSize));

  for (let index = 0; index < totalChunks; index += 1) {
    const start = index * chunkSize;
    const end = Math.min(file.size, start + chunkSize);
    const chunk = await file.slice(start, end).arrayBuffer();
    spark.append(chunk);
    onProgress?.((index + 1) / totalChunks);
  }

  return spark.end();
}
