export function captureIo() {
  const chunks = [];
  const io = {
    log(message) {
      chunks.push(String(message));
    },
    error(message) {
      chunks.push(String(message));
    },
  };
  return {
    io,
    output() {
      return chunks.join("\n");
    },
  };
}
