package core

import (
	"errors"

	"github.com/gosimple/slug"
	"github.com/mitchellh/mapstructure"
)

// Product is struct that contains product info
type Product struct {
	ProductID   int    `db:"product_id"`
	Name        string `db:"name"`
	Price       int    `db:"price"`
	Discount    int    `db:"discount"`
	Quantity    int    `db:"quantity"`
	Description string `db:"description"`
	URL         string `db:"url"`
	Recommended bool   `db:"recommended"`
	CreatedAt   string `db:"created_at"`
	UpdatedAt   string `db:"updated_at"`
	Brand       string `db:"brand"`
	BrandID     int    `db:"brand_id"`
	Category    string `db:"category"`
	CategoryID  int    `db:"category_id"`
	Images      []struct {
		ImageID int    `db:"product_image_id"`
		URL     string `db:"url"`
	}
	Features []struct {
		FeatureID      int    `db:"product_feature_id"`
		FeatureValueID int    `db:"product_feature_value_id"`
		Name           string `db:"name"`
		Value          string `db:"value"`
		Recommended    bool   `db:"recommended"`
	}
	Publications []struct {
		PublicationID int    `db:"publication_id"`
		Name          string `db:"name"`
	}
	Recommendations []Product
}

// ProductGet returns single product
func ProductGet(productID int, light bool) (Product, error) {
	product := Product{}

	// get light columns
	err := Global.DB.Get(&product,
		`SELECT product_id, p.name, price, discount, quantity, p.url, recommended, updated_at,
			created_at, brand_id, b.name AS brand, category_id, c.name AS category
		FROM products p
		INNER JOIN brands b USING (brand_id)
		INNER JOIN categories c USING (category_id)
		WHERE product_id = $1;`, productID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// get heavy columns
	if !light {
		err := Global.DB.Get(&product,
			`SELECT description
			FROM products
			WHERE product_id = $1;`, productID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}

		err = Global.DB.Select(&product.Images,
			`SELECT product_image_id, url
			FROM product_images
			WHERE product_id = $1;`, productID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}

		err = Global.DB.Select(&product.Features,
			`SELECT product_feature_id, product_feature_value_id, name, value, recommended
			FROM product_feature_values
			INNER JOIN product_features USING (product_feature_id)
			WHERE product_id = $1;`, productID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}

		err = Global.DB.Select(&product.Publications,
			`SELECT publication_id, name
			FROM product_publications
			INNER JOIN publications USING (publication_id)
			WHERE product_id = $1;`, productID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}

		err = Global.DB.Select(&product.Recommendations,
			`SELECT p.product_id, p.name, price, discount, quantity, p.url, recommended, updated_at,
				created_at, brand_id, b.name AS brand, category_id, c.name AS category
			FROM product_recommendations r
			INNER JOIN products p ON recommended_product_id = p.product_id
			INNER JOIN brands b USING (brand_id)
			INNER JOIN categories c USING (category_id)
			WHERE r.product_id = $1;`, productID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	return product, nil
}

// ProductList returns list of products
func ProductList(offset int, limit int) ([]Product, error) {
	products := []Product{}

	err := Global.DB.Select(&products,
		`SELECT product_id, p.name, price, discount, quantity, p.url, recommended, updated_at,
			created_at, brand_id, b.name AS brand, category_id, c.name AS category
		FROM products p
		INNER JOIN brands b USING (brand_id)
		INNER JOIN categories c USING (category_id)
		ORDER BY updated_at DESC
		OFFSET $1 LIMIT $2;`, offset, limit)
	if err != nil {
		Global.Log.Error(err)
		return []Product{}, err
	}

	return products, nil
}

// ProductSearch returns list of products filtered by query
func ProductSearch(query string, offset int, limit int) ([]Product, error) {
	products := []Product{}

	err := Global.DB.Select(&products,
		`SELECT product_id, ts_headline(p.name, plainto_tsquery(unaccent($1))) AS name,
			price, discount, quantity, p.url, recommended, updated_at, created_at,
			brand_id, b.name AS brand, category_id, c.name AS category
		FROM products p
		INNER JOIN brands b USING (brand_id)
		INNER JOIN categories c USING (category_id)
		WHERE tsv @@ plainto_tsquery(unaccent($1))
		ORDER BY ts_rank(tsv, plainto_tsquery(unaccent($1))) DESC
		OFFSET $2 LIMIT $3;`, query, offset, limit)
	if err != nil {
		Global.Log.Error(err)
		return []Product{}, err
	}

	return products, nil
}

// ProductCheck returns true if there are created/updated products
func ProductCheck(since string) (bool, error) {
	var modified bool
	err := Global.DB.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM products
			WHERE updated_at > $1
			LIMIT 1
		);`, since).Scan(&modified)
	if err != nil {
		Global.Log.Error(err)
		return false, err
	}
	return modified, nil
}

// ProductCreate creates product in db and returns it
func ProductCreate(data map[string]interface{}) (Product, error) {
	product := Product{}

	// decode request parameters
	err := mapstructure.Decode(data, &product)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// check if product is valid
	if product.Name == "" {
		err := errors.New("Proizvod mora imati ime")
		Global.Log.Error(err)
		return Product{}, err
	}
	if product.Price == 0 {
		err := errors.New("Proizvod mora imati cijenu")
		Global.Log.Error(err)
		return Product{}, err
	}
	if product.Brand == "" {
		err := errors.New("Proizvod mora imati brend")
		Global.Log.Error(err)
		return Product{}, err
	}
	if product.Category == "" {
		err := errors.New("Proizvod mora imati kategoriju")
		Global.Log.Error(err)
		return Product{}, err
	}

	// create product url from its name
	product.URL = slug.Make(product.Name)

	// start db transaction
	tx, err := Global.DB.Begin()
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// if brand has a name but invalid id, it should be created
	if product.BrandID == 0 {
		brand, err := BrandCreate(map[string]interface{}{
			"Name": product.Brand,
		})
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
		product.BrandID = brand.BrandID
	}

	// if category has a name but invalid id, it should be created
	if product.CategoryID == 0 {
		category, err := CategoryCreate(map[string]interface{}{
			"Name": product.Category,
		})
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
		product.CategoryID = category.CategoryID
	}

	// insert product
	err = tx.QueryRow(
		`INSERT INTO products (name, price, discount, quantity, description, url,
			recommended, created_at, updated_at, brand_id, category_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW()::TIMESTAMP, NOW()::TIMESTAMP, $8, $9)
		RETURNING product_id;`, product.Name, product.Price, product.Discount,
		product.Quantity, product.Description, product.URL, product.Recommended,
		product.BrandID, product.CategoryID).Scan(&product.ProductID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// insert images
	for _, image := range product.Images {
		_, err := tx.Exec(
			`INSERT INTO product_images (url, product_id)
			VALUES ($1, $2);`, image.URL, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// insert feature values
	for _, feature := range product.Features {
		// if feature has invalid name, it should be created
		if feature.FeatureID == 0 {
			err := tx.QueryRow(
				`INSERT INTO product_features (name, recommended, category_id)
				VALUES ($1, $2, $3)
				RETURNING product_feature_id;`, feature.Name, feature.Recommended,
				product.CategoryID).Scan(&feature.FeatureID)
			if err != nil {
				Global.Log.Error(err)
				return Product{}, err
			}
		}

		_, err := tx.Exec(
			`INSERT INTO product_feature_values (value, product_feature_id, product_id)
			VALUES ($1, $2, $3);`, feature.Value, feature.FeatureID, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// insert publications
	for _, publication := range product.Publications {
		_, err := tx.Exec(
			`INSERT INTO product_publications (publication_id, product_id)
			VALUES ($1, $2);`, publication.PublicationID, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// insert recommendations
	for _, recommendation := range product.Recommendations {
		_, err := tx.Exec(
			`INSERT INTO product_recommendations (recommended_product_id, product_id)
			VALUES ($1, $2);`, recommendation.ProductID, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// get created product
	createdproduct, err := ProductGet(product.ProductID, false)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	return createdproduct, nil
}

// ProductUpdate updates product in db and returns it
func ProductUpdate(data map[string]interface{}) (Product, error) {
	product := Product{}

	// decode request parameters
	err := mapstructure.Decode(data, &product)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// check if product is valid
	if product.Name == "" {
		err := errors.New("Proizvod mora imati ime")
		Global.Log.Error(err)
		return Product{}, err
	}
	if product.Price == 0 {
		err := errors.New("Proizvod mora imati cijenu")
		Global.Log.Error(err)
		return Product{}, err
	}
	if product.Brand == "" {
		err := errors.New("Proizvod mora imati brend")
		Global.Log.Error(err)
		return Product{}, err
	}
	if product.Category == "" {
		err := errors.New("Proizvod mora imati kategoriju")
		Global.Log.Error(err)
		return Product{}, err
	}

	// create product url from its name
	product.URL = slug.Make(product.Name)

	// start db transaction
	tx, err := Global.DB.Begin()
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// check if product exists
	var exists bool
	err = tx.QueryRow(
		`SELECT EXISTS(
			SELECT 1
			FROM products
			WHERE product_id = $1
		);`, product.ProductID).Scan(&exists)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}
	if !exists {
		err := errors.New("Proizvod ne postoji")
		Global.Log.Error(err)
		return Product{}, err
	}

	// if brand has a name but invalid id, it should be created
	if product.BrandID == 0 {
		brand, err := BrandCreate(map[string]interface{}{
			"Name": product.Brand,
		})
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
		product.BrandID = brand.BrandID
	}

	// if category has a name but invalid id, it should be created
	if product.CategoryID == 0 {
		category, err := CategoryCreate(map[string]interface{}{
			"Name": product.Category,
		})
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
		product.CategoryID = category.CategoryID
	}

	// update product
	_, err = tx.Exec(
		`UPDATE products
		SET name = $2,
			price = $3,
			discount = $4,
			quantity = $5,
			description = $6,
			url = $7,
			recommended = $8,
			updated_at = NOW()::TIMESTAMP,
			brand_id = $9,
			category_id = $10
		WHERE product_id = $1;`, product.ProductID, product.Name, product.Price,
		product.Discount, product.Quantity, product.Description, product.URL,
		product.Recommended, product.BrandID, product.CategoryID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// delete old images
	_, err = tx.Exec(
		`DELETE FROM product_images
		WHERE product_id = $1;`, product.ProductID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// insert new images
	for _, image := range product.Images {
		_, err := tx.Exec(
			`INSERT INTO product_images (url, product_id)
			VALUES ($1, $2);`, image.URL, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// delete old feature values
	_, err = tx.Exec(
		`DELETE FROM product_feature_values
		WHERE product_id = $1;`, product.ProductID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// insert new feature values
	for _, feature := range product.Features {
		// if feature has invalid name, it should be created
		if feature.FeatureID == 0 {
			err := tx.QueryRow(
				`INSERT INTO product_features (name, recommended, category_id)
				VALUES ($1, $2, $3)
				RETURNING product_feature_id;`, feature.Name, feature.Recommended,
				product.CategoryID).Scan(&feature.FeatureID)
			if err != nil {
				Global.Log.Error(err)
				return Product{}, err
			}
		}

		_, err := tx.Exec(
			`INSERT INTO product_feature_values (value, product_feature_id, product_id)
			VALUES ($1, $2, $3);`, feature.Value, feature.FeatureID, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// delete old publications
	_, err = tx.Exec(
		`DELETE FROM product_publications
		WHERE product_id = $1;`, product.ProductID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// insert new publications
	for _, publication := range product.Publications {
		_, err := tx.Exec(
			`INSERT INTO product_publications (publication_id, product_id)
			VALUES ($1, $2);`, publication.PublicationID, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// delete old recommendations
	_, err = tx.Exec(
		`DELETE FROM product_recommendations
		WHERE product_id = $1;`, product.ProductID)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// insert new recommendations
	for _, recommendation := range product.Recommendations {
		_, err := tx.Exec(
			`INSERT INTO product_recommendations (recommended_product_id, product_id)
			VALUES ($1, $2);`, recommendation.ProductID, product.ProductID)
		if err != nil {
			Global.Log.Error(err)
			return Product{}, err
		}
	}

	// commit transaction
	err = tx.Commit()
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	// get updated product
	updatedproduct, err := ProductGet(product.ProductID, false)
	if err != nil {
		Global.Log.Error(err)
		return Product{}, err
	}

	return updatedproduct, nil
}