## **File Structure:**
- **cmd**: go entrypoints
    - **prep**: runs during ci. creates JSON graph spec, packages vault files to sqllite database
    - **server**: web server
- **internal**: server app private packages
    - **handler**: route handler functions
    - **render**: go html template renderer
- **pkg**: exported packages
    - **filegraph**: builds JSON graph file from vault
- **public**: static content
    - **css**
    - **views**: go html templates
- **main.ts**: initializes sigmaJS graph on DOM load


## **Environment Setup**
0. Requires Node >= 14.0.0, Go >= 1.23.0

1. Install dependencies

    `npm install chroma-js graphology graphology-layout-force sigma uuid`
    
    `go mod tidy`

2. Place vault folder at /vault

## **Local Build and Run**
1. Create graph file

    `go run cmd/prep/main.go`
2. Bundle static content and JS

    `npx vite build`
3. Start server

    `go run cmd/server/main.go`