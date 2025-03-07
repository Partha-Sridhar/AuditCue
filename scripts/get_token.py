#!/usr/bin/env python3
"""
Script to get OAuth tokens for Gmail.

This script uses Google's OAuth2 flow to get an authentication token
for Gmail API access. Run this script to generate a refresh token
that can be used by the integration service.
"""

from google_auth_oauthlib.flow import InstalledAppFlow

def main():
    """Run the OAuth flow to get Gmail tokens."""
    print("Starting Gmail OAuth token generator...")
    
    # Define the scopes needed
    SCOPES = ['https://www.googleapis.com/auth/gmail.send']
    
    # Create a flow instance with client secrets file
    print("Loading credentials from 'credentials.json'...")
    try:
        flow = InstalledAppFlow.from_client_secrets_file(
            'credentials.json',
            scopes=SCOPES
        )
    except Exception as e:
        print(f"Error loading credentials: {e}")
        print("Make sure 'credentials.json' exists in the current directory.")
        return
    
    # Run the flow to get credentials
    print("Starting local server for authentication...")
    print("Your browser will open to complete the authentication process.")
    creds = flow.run_local_server(port=8080)
    
    # Print the tokens
    print("\n--- OAuth Tokens ---")
    print(f"Access Token: {creds.token}")
    print(f"Refresh Token: {creds.refresh_token}")
    print(f"Expires at: {creds.expiry}")
    print("\nIMPORTANT: Save the refresh token! You'll need it for the integration service.")

if __name__ == "__main__":
    main()