// Quest id from a yrv query object. Duplicate keys arrive as an array.
export function questIdFromQuery(query) {
  const raw = query && query.id;
  if (Array.isArray(raw)) return raw[0] || "";
  return typeof raw === "string" ? raw : "";
}
