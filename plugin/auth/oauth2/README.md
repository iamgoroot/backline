# Backline oauth2 plugin

## Example config for google

oauth2:
  enabled: true
  clientId: "client-id"
  clientSecret: "secret"
  redirectUrl: "http://localhost:8080/oauth2/callback"
  scopes:
    - "https://www.googleapis.com/auth/userinfo.email"
  endpoint:
    authUrl: "https://accounts.google.com/o/oauth2/auth"
    tokenUrl: "https://accounts.google.com/o/oauth2/token"
  userInfo:
    bodyDecoder: base64
    url: "https://www.googleapis.com/oauth2/v3/tokeninfo"
    emailPath: "email"
    emailVerifiedPath: "email_verified"