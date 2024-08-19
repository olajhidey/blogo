package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"

	firebase "firebase.google.com/go/v4"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/api/option"
)

var database *sql.DB

type Blog struct {
	Id          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	BlogImage   string `json:"blog_image"`
	DateCreated string `json:"date_created"`
	Content     string `json:"content"`
	TagId       string `json:"tag_id"`
}

type RequestBlog struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	BlogImage   string `json:"blog_image,omitempty"`
	Content     string `json:"content"`
	TagId       string `json:"tag_id"`
}

type Tag struct {
	Id   string `json:"id,omitempty"`
	Name string `json:"name"`
}

// create a blog
// add content to the blog
// title, description, tag or category, image (if necessary)

// Edit a blog
// path in UID and update the blog item

// Delete a blog
// path in UID and create a method to delete the blog

// Create a tag or collection
// endpoint to create a list of tag and collection

// List all of the blog itens
// Endpoint to provide list of blog items for an authenticated user

// You need a function to upload image to S3
// another function to generate presigned url from S3

func loadDb() *sql.DB {
	dsn := "blogo:blogo@/blogo"
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		fmt.Println("Error opening database: ", err)
	}

	return db
}

func createTables(database *sql.DB) {
	err := database.Ping()
	if err != nil {
		fmt.Println("Error pinging database: ", err)
	}

	fmt.Println("Connected to the database successfully!")

	if err := PreloadSchema(database); err != nil {
		fmt.Println(err)
	}
}

func GetAllBlogs(ctx *gin.Context) {

	blogs := "SELECT id, title, description,tag_id, blog_image, date_created, content FROM blog"
	rows, err := database.Query(blogs)

	if err != nil {
		fmt.Println("Error Executing query: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	defer rows.Close()

	var results []Blog

	for rows.Next() {
		var id int
		var title string
		var description string
		var tag_id string
		var blog_image string
		var date_created string
		var content string

		if err := rows.Scan(&id, &title, &description, &tag_id, &blog_image, &date_created, &content); err != nil {
			fmt.Println("Error scanning row: ", err)
			return
		}

		results = append(results, Blog{
			Id:          id,
			Title:       title,
			Description: description,
			BlogImage:   blog_image,
			DateCreated: date_created,
			Content:     content,
			TagId:       tag_id,
		})
	}

	ctx.JSON(http.StatusOK, results)

}

func GetBlog(ctx *gin.Context) {

	var blog Blog

	id := ctx.Param("id")

	query := "SELECT id, title, description, blog_image, content, date_created, tag_id FROM blog WHERE id = ?"

	result := database.QueryRow(query, id)

	err := result.Scan(&blog.Id, &blog.Title, &blog.Description, &blog.BlogImage, &blog.Content, &blog.DateCreated, &blog.TagId)

	if err != nil {

		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Blog not Found",
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, &blog)
}

func CreateBlog(ctx *gin.Context) {

	var requestBody RequestBlog

	if err := ctx.ShouldBindBodyWithJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	insertBlog := `INSERT INTO blog (title, description, tag_id, blog_image, date_created, content) VALUES (?,?,?,?,?,?)`

	statement, err := database.Prepare(insertBlog)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	dateCreated := time.Now()

	result, err := statement.Exec(requestBody.Title, requestBody.Description, requestBody.TagId, requestBody.BlogImage, dateCreated, requestBody.Content)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	lastInserted, err := result.LastInsertId()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	fmt.Printf("Last inserted: %d\n", lastInserted)

	fmt.Println("Todo Inserted Successfully!!")

	ctx.JSON(http.StatusOK, gin.H{
		"id":           lastInserted,
		"title":        requestBody.Title,
		"description":  requestBody.Description,
		"content":      requestBody.Content,
		"date_created": dateCreated,
		"blog_image":   requestBody.BlogImage,
		"tag_id":       requestBody.TagId,
	})

}

func UpdateBlog(ctx *gin.Context) {
	id := ctx.Param("id")

	var blog RequestBlog

	if err := ctx.ShouldBindBodyWithJSON(&blog); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	updateQuery := "UPDATE blog SET title = ?, description = ?, blog_image = ?, content = ?, tag_id = ? WHERE id = ?"
	result, err := database.Exec(updateQuery, blog.Title, blog.Description, blog.BlogImage, blog.Content, blog.TagId, id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Blog not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Todo Updated successfully!",
	})

}

func DeleteBlog(ctx *gin.Context) {

	id := ctx.Param("id")
	deleteQuery := "DELETE FROM blog WHERE id = ?"
	result, err := database.Exec(deleteQuery, id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Blog not Found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Blog deleted Successfully!",
	})
}

// Tags

func GetTags(ctx *gin.Context) {

	var tags []Tag

	tagsQuery := "SELECT id, name FROM tag"
	result, err := database.Query(tagsQuery)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	for result.Next() {
		var id string
		var name string

		if err := result.Scan(&id, &name); err != nil {
			fmt.Printf("Error scanning row: %v", err)
		}

		tags = append(tags, Tag{
			Id:   id,
			Name: name,
		})
	}

	ctx.JSON(http.StatusOK, tags)

}

func DeleteTag(ctx *gin.Context) {
	id := ctx.Param("id")

	deleteQuery := "DELETE FROM tag WHERE id = ?"
	result, err := database.Exec(deleteQuery, id)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if rowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "Item not found",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Tag deleted successfully!",
	})

}

func CreateTag(ctx *gin.Context) {

	var tag Tag

	if err := ctx.ShouldBindBodyWithJSON(&tag); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	insertTag := "INSERT INTO tag (name) VALUES (?)"
	statement, err := database.Prepare(insertTag)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	result, err := statement.Exec(tag.Name)

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	fmt.Println("Tag added successfully!")

	lastInserted, err := result.LastInsertId()

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	fmt.Printf("Last Inserted Id: %d", lastInserted)

	ctx.JSON(http.StatusOK, gin.H{
		"id":   lastInserted,
		"name": tag.Name,
	})
}

func GetTag(ctx *gin.Context){
	id := ctx.Param("id")
	var tag Tag
	query := "SELECT name FROM tag WHERE id = ?" 
	result := database.QueryRow(query, id)

	if err := result.Scan(&tag.Name); err != nil{
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Tag not found",
			})
			return 
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return 
	}

	ctx.JSON(http.StatusOK, gin.H{
		"name": tag.Name,
	})
}

func main() {

	router := gin.Default()

	config := &firebase.Config{
		StorageBucket: "firestoreapp-87f7f.appspot.com",
	}

	opt := option.WithCredentialsFile("./credentials.json")

	app, err := firebase.NewApp(context.Background(), config, opt)

	if err != nil {
		log.Fatalln(err)
	}

	router.Static("/static", "./www")

	router.GET("/", func(ctx *gin.Context) {
		ctx.File("./www/index.html")
	})

	api := router.Group("/")

	database = loadDb()

	createTables(database)

	defer database.Close()

	// Get all Blogs
	api.GET("/api/blogs", GetAllBlogs)

	// Get a selected Blog item
	api.GET("/api/blog/:id", GetBlog)

	// Create Blog
	api.POST("/api/blog/create", CreateBlog)

	// Update Blog
	api.PUT("/api/blog/update/:id", UpdateBlog)

	// Delete a Blog item
	api.DELETE("/api/blog/delete/:id", DeleteBlog)

	// Tags or categories
	// Get all Tags or categories
	api.GET("/api/tags", GetTags)

	// Delete a tag
	api.DELETE("/api/tag/delete/:id", DeleteTag)

	// Create a new Tag
	api.POST("/api/tag/create", CreateTag)

	api.GET("/api/tag/:id", GetTag)

	api.POST("/upload", func(ctx *gin.Context) {

		file, _ := ctx.FormFile("file")

		log.Printf("File Size: %v\n", file.Size)

		src, err := file.Open()

		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
		}
		defer src.Close()

		imageUrl, err := uploadToFirebase(app, src, time.Now().UTC().Format("20060102150405"))

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return 
		}

		ctx.JSON(http.StatusOK, gin.H{
			"url": imageUrl,
		})

	})

	router.Run(":8080")

}

func uploadToFirebase(app *firebase.App, file multipart.File, filename string)(string, error){

	ctx := context.Background()

	client, err := app.Storage(ctx)
	if err != nil {
		return "", fmt.Errorf("Error getting storage client: %v", err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return "", fmt.Errorf("Error getting default bucket: %v", err)
	}

	storageWriter := bucket.Object("uploads/"+filename).NewWriter(ctx)

	// storageWriter.ContentType = "image/jpeg"

	if _,err := file.Seek(0,0); err != nil{
		return "", err
	}

	if _, err := io.Copy(storageWriter, file); err != nil {
		return "", err
	}

	// Close the writer
	if err := storageWriter.Close(); err != nil {
		return "", err
	}

	attrs, err := bucket.Object("uploads/"+filename).Attrs(ctx)

	if err != nil {
		return "", err
	}

	log.Printf("Uploaded file size: %v\n", attrs.Size)

	baseURL := "https://firebasestorage.googleapis.com/v0/b/" + attrs.Bucket + "/o/"
	encodedObjectName := url.PathEscape(attrs.Name)
	imageURL := baseURL + encodedObjectName + "?alt=media"
	return imageURL, nil

}
