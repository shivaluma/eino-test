import { NextRequest, NextResponse } from 'next/server';

export async function GET(request: NextRequest) {
  try {
    // Check if access token cookie exists
    const accessToken = request.cookies.get('access_token')?.value;
    
    if (!accessToken) {
      return NextResponse.json({ authenticated: false }, { status: 401 });
    }

    // Token exists in HTTP-only cookie, user is authenticated
    return NextResponse.json({ 
      authenticated: true,
      // Don't return the actual token for security
      hasToken: true 
    });
    
  } catch (error) {
    console.error('Token validation error:', error);
    return NextResponse.json({ authenticated: false }, { status: 401 });
  }
}