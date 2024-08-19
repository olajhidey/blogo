package main

import (
	"database/sql"
	"fmt"
)

func PreloadSchema(db *sql.DB) error{
	tagSchema := `
	CREATE TABLE IF NOT EXISTS tag (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(100) NOT NULL
    );
	`

	_, err := db.Exec(tagSchema)

	if err != nil {
		return fmt.Errorf("Error creating schema: %v", err)
	}

	blogSchema := `
	CREATE TABLE IF NOT EXISTS blog (
        id INT AUTO_INCREMENT PRIMARY KEY,
        title VARCHAR(255) NOT NULL,
        description TEXT NOT NULL,
        tag_id INT NOT NULL,
        blog_image VARCHAR(255),
        date_created DATETIME DEFAULT CURRENT_TIMESTAMP,
        content TEXT NOT NULL,
        FOREIGN KEY (tag_id) REFERENCES tag(id) ON DELETE CASCADE
    );
	`
	_ , err = db.Exec(blogSchema)

	if err != nil {
		return fmt.Errorf("Error creating blog schema: %v", err)
	}

	fmt.Println("Schema preloaded successfully!")
	return nil
}