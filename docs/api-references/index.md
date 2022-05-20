# API-Referens

## POST /v1/users
Skapa en ny användare.
```
// Request
{
    "display_name": "<public_display_name>",
    "email": "<name>@<domain>",
    "password": "<hmac256_encrypted_password>"
    "accepted_terms": true,
    "icon_uuid": "<icon_uuid>"
}

// Response
Returnerar ett User-objekt.
```

