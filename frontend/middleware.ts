import { NextResponse, type NextRequest } from "next/server";

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  /*
   * Playwright starts the dev server and requires a 200 status to
   * begin the tests, so this ensures that the tests can start
   */
  if (pathname === "/api/health") {
    return NextResponse.next();
  }

  // Skip middleware for OAuth callback routes and auth pages
  const publicRoutes = [
    "/sign-in",
    "/sign-up", 
    "/oauth",
    "/callback",
    "/api/auth"
  ];
  
  const isPublicRoute = publicRoutes.some(route => pathname.startsWith(route));
  if (isPublicRoute) {
    return NextResponse.next();
  }

  // Check for authentication token in cookies or localStorage
  // Since middleware runs on server side, we need to check cookies
  const token = request.cookies.get('access_token')?.value;
  
  if (!token) {
    // Redirect to sign-in if not authenticated
    return NextResponse.redirect(new URL("/sign-in", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    "/((?!_next/static|_next/image|favicon.ico|sitemap.xml|robots.txt|api/auth|sign-in|sign-up|oauth|callback).*)",
  ],
};
