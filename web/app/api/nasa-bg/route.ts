// Returns the latest NASA "Image of the Day" — the first item's enclosure URL
// from the public IOTD RSS feed. Fetched server-side (the feed sends no CORS
// headers); the image itself is then loaded directly by the browser. Cached for
// a few hours since the feed updates roughly daily.
const FEED = "https://www.nasa.gov/feeds/iotd-feed/";

export const revalidate = 10800; // 3h

export async function GET() {
  try {
    const res = await fetch(FEED, {
      headers: { "User-Agent": "SlashNode" },
      next: { revalidate },
    });
    if (!res.ok) {
      return Response.json({ error: "feed unavailable" }, { status: 502 });
    }
    const xml = await res.text();
    const firstItem = xml.split(/<item>/i)[1] ?? "";
    const url = firstItem.match(
      /<enclosure[^>]*\burl="([^"]+)"/i,
    )?.[1];
    const title = decodeXml(firstItem.match(/<title>([\s\S]*?)<\/title>/i)?.[1] ?? "");
    if (!url) {
      return Response.json({ error: "no image in feed" }, { status: 502 });
    }
    return Response.json({ url, title });
  } catch {
    return Response.json({ error: "feed fetch failed" }, { status: 502 });
  }
}

function decodeXml(s: string): string {
  return s
    .replace(/<!\[CDATA\[([\s\S]*?)\]\]>/g, "$1")
    .replace(/&amp;/g, "&")
    .replace(/&quot;/g, '"')
    .replace(/&#0?39;|&apos;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">")
    .trim();
}
