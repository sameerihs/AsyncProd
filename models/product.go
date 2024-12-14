package models

import (
	"AsyncProd/config"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lib/pq"
)


type Product struct {
	ID                  int       `json:"id"`
	UserID              int       `json:"user_id"`
	ProductName         string    `json:"product_name"`
	ProductDescription  string    `json:"product_description"`
	ProductImages       []string  `json:"product_images"`
	ProductPrice        float64   `json:"product_price"`
	CompressedImages    []string  `json:"compressed_product_images,omitempty"`
	CreatedAt           time.Time `json:"created_at,omitempty"`
	UpdatedAt           time.Time `json:"updated_at,omitempty"`
}

func (p *Product) Validate() error {
	if p.UserID <= 0 {
		return errors.New("invalid user ID")
	}
	if p.ProductName == "" {
		return errors.New("product name is required")
	}
	if p.ProductPrice < 0 {
		return errors.New("product price cannot be negative")
	}
	return nil
}


func SaveProduct(product *Product) (int, error) {

    if err := product.Validate(); err != nil {
        return 0, err
    }

    query := `
        INSERT INTO products (
            user_id, 
            product_name, 
            product_description, 
            product_price, 
            product_images,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
        RETURNING id
    `
    var productID int
    err := config.DB.QueryRow(
        query, 
        product.UserID, 
        product.ProductName, 
        product.ProductDescription, 
        product.ProductPrice, 
        pq.Array(product.ProductImages),
    ).Scan(&productID)

    if err != nil {
        
        log.Printf("Failed to save product: %v", err)
        return 0, fmt.Errorf("failed to save product: %v", err)
    }

    return productID, nil
}

func GetProductByID(id int) (*Product, error) {
	var product Product
	var images pq.StringArray
	var compressedImages pq.StringArray

	query := `
		SELECT 
			id, 
			user_id, 
			product_name, 
			product_description, 
			product_price, 
			product_images, 
			compressed_product_images,
			created_at,
			updated_at
		FROM products
		WHERE id = $1
	`
	err := config.DB.QueryRow(query, id).Scan(
		&product.ID, 
		&product.UserID, 
		&product.ProductName, 
		&product.ProductDescription,
		&product.ProductPrice, 
		&images, 
		&compressedImages,
		&product.CreatedAt,
		&product.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("product not found")
		}
		return nil, fmt.Errorf("error retrieving product: %v", err)
	}

	product.ProductImages = images
	product.CompressedImages = compressedImages
	return &product, nil
}


func UpdateProduct(product *Product) error {
	if err := product.Validate(); err != nil {
		return err
	}

	query := `
		UPDATE products
		SET 
			product_name = $2, 
			product_description = $3, 
			product_price = $4, 
			product_images = $5,
			compressed_product_images = $6,
			updated_at = NOW()
		WHERE id = $1 AND user_id = $7
	`
	result, err := config.DB.Exec(
		query,
		product.ID,
		product.ProductName,
		product.ProductDescription,
		product.ProductPrice,
		pq.Array(product.ProductImages),
		pq.Array(product.CompressedImages),
		product.UserID,
	)

	if err != nil {
		return fmt.Errorf("failed to update product: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking update result: %v", err)
	}

	if rowsAffected == 0 {
		return errors.New("no product found or unauthorized to update")
	}

	return nil
}

func GetProductsByUserID(userID int, minPrice, maxPrice float64, productName string) ([]Product, error) {
	query := `
		SELECT 
			id, 
			user_id, 
			product_name, 
			product_description, 
			product_price, 
			product_images, 
			compressed_product_images,
			created_at,
			updated_at
		FROM products
		WHERE user_id = $1
			AND ($2 = 0 OR product_price BETWEEN $2 AND $3)
			AND ($4 = '' OR product_name ILIKE $4)
		ORDER BY created_at DESC
	`
	
	rows, err := config.DB.Query(query, userID, minPrice, maxPrice, "%"+productName+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve products: %v", err)
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		var images, compressedImages pq.StringArray
		
		err := rows.Scan(
			&product.ID, 
			&product.UserID, 
			&product.ProductName, 
			&product.ProductDescription,
			&product.ProductPrice, 
			&images, 
			&compressedImages,
			&product.CreatedAt,
			&product.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning product: %v", err)
		}

		product.ProductImages = images
		product.CompressedImages = compressedImages
		products = append(products, product)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during product retrieval: %v", err)
	}

	return products, nil
}
