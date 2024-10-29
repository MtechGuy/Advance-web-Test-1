CREATE TABLE products (
    product_id bigserial PRIMARY KEY,
    name text NOT NULL,
    description text NOT NULL,
    category text NOT NULL,
    image_url text NOT NULL,
    average_rating DECIMAL(3, 2) DEFAULT 0.00, -- Average rating from reviews
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version integer NOT NULL DEFAULT 1
);

-- Step 2: Create the Reviews Table
CREATE TABLE reviews (
    review_id bigserial PRIMARY KEY,
    product_id INT REFERENCES products(product_id) ON DELETE CASCADE,
    rating INT CHECK (rating BETWEEN 1 AND 5),
    review_text text NOT NULL,
    helpful_count INT DEFAULT 0,
    created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
     version integer NOT NULL DEFAULT 1
);