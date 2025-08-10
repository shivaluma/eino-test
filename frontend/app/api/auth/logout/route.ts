import { NextRequest, NextResponse } from 'next/server';

export async function POST(_request: NextRequest) {
  try {
    // Create response with cleared cookies
    const response = NextResponse.json({ success: true, message: 'Logged out successfully' });

    // Clear the access token cookie
    response.cookies.set('access_token', '', {
      path: '/',
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      maxAge: 0, // Expire immediately
    });

    // Clear the refresh token cookie
    response.cookies.set('refresh_token', '', {
      path: '/',
      httpOnly: true,
      secure: true,
      sameSite: 'lax',
      maxAge: 0, // Expire immediately
    });

    return response;
  } catch (error) {
    console.error('Logout error:', error);
    return NextResponse.json({ success: false, error: 'Logout failed' }, { status: 500 });
  }
}