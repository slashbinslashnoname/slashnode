// Clears the session cookie, ending the operator's session. The middleware then
// redirects to /login on the next protected request.
export async function POST() {
  const response = Response.json({ status: "ok" });
  response.headers.append(
    "Set-Cookie",
    "slashnode_session=; HttpOnly; Path=/; SameSite=Lax; Max-Age=0",
  );
  return response;
}
