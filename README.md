# Blogo

This project is a representation of using Golang to build a simple Blog Web application. The application is built using the following:

- MySQL (for database)
- HTML, CSS and Vue (Frontend)
- Golang (Backend)
- [Firebase Storage](https://firebase.google.com/docs/storage) - This is used to save the blog preview URL 


## Run this application 
- Clone this repository
- Change `.env.example` file to `.env` and provide the necessary credentials.
- Get your [serviceKey credentials on the firebase console](https://firebase.google.com/docs/admin/setup) and add it to the root folder of this project.
- Make sure you have MySQL setup and reference the credentials in the `loadDb` method in the `main.go` file.
- Run the application using `go run main.go schema.go` (We need the schema.go for our preloaded Schema)