-- public.products definition

-- Drop table

-- DROP TABLE products;

CREATE TABLE products (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	sku varchar(50) NOT NULL,
	"name" varchar(255) NOT NULL,
	price numeric(10, 2) NOT NULL,
	inventory int4 DEFAULT 0 NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT products_pkey PRIMARY KEY (id),
	CONSTRAINT products_sku_key UNIQUE (sku)
);


-- public.promotions definition

-- Drop table

-- DROP TABLE promotions;

CREATE TABLE promotions (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	"type" varchar(50) NOT NULL,
	description text NOT NULL,
	"rule" jsonb NOT NULL,
	active bool DEFAULT true NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT promotions_pkey PRIMARY KEY (id)
);


-- public.users definition

-- Drop table

-- DROP TABLE users;

CREATE TABLE users (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	email varchar(255) NOT NULL,
	"password" varchar(255) NOT NULL,
	"name" varchar(255) NOT NULL,
	"role" varchar(20) DEFAULT 'customer'::character varying NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	deleted_at timestamptz NULL,
	CONSTRAINT users_email_key UNIQUE (email),
	CONSTRAINT users_pkey PRIMARY KEY (id)
);


-- public.cart_items definition

-- Drop table

-- DROP TABLE cart_items;

CREATE TABLE cart_items (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	product_id uuid NOT NULL,
	quantity int4 DEFAULT 1 NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	deleted_at timestamptz NULL,
	user_id uuid NULL, -- User who owns this cart item
	CONSTRAINT cart_items_pkey PRIMARY KEY (id),
	CONSTRAINT cart_items_user_id_product_id_key UNIQUE (user_id, product_id),
	CONSTRAINT cart_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
	CONSTRAINT cart_items_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_cart_items_user_id ON public.cart_items USING btree (user_id);
COMMENT ON TABLE public.cart_items IS 'Stores items in a user''s cart with direct reference to products';

-- Column comments

COMMENT ON COLUMN public.cart_items.user_id IS 'User who owns this cart item';


-- public.checkouts definition

-- Drop table

-- DROP TABLE checkouts;

CREATE TABLE checkouts (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	subtotal numeric(10, 2) NOT NULL,
	total_discount numeric(10, 2) NOT NULL,
	total numeric(10, 2) NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	user_id uuid NULL,
	payment_status varchar(50) DEFAULT 'PENDING'::character varying NOT NULL, -- Payment status: PENDING, PAID, FAILED, REFUNDED
	payment_method varchar(50) NULL,
	payment_reference varchar(255) NULL,
	notes text NULL,
	status varchar(50) DEFAULT 'CREATED'::character varying NOT NULL, -- Order status: CREATED, PROCESSING, SHIPPED, DELIVERED, CANCELLED
	completed_at timestamptz NULL,
	CONSTRAINT checkouts_pkey PRIMARY KEY (id),
	CONSTRAINT checkouts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id)
);
CREATE INDEX idx_checkouts_payment_status ON public.checkouts USING btree (payment_status);
CREATE INDEX idx_checkouts_status ON public.checkouts USING btree (status);
CREATE INDEX idx_checkouts_user_id ON public.checkouts USING btree (user_id);
COMMENT ON TABLE public.checkouts IS 'Stores order information, representing completed checkouts with payment status';

-- Column comments

COMMENT ON COLUMN public.checkouts.payment_status IS 'Payment status: PENDING, PAID, FAILED, REFUNDED';
COMMENT ON COLUMN public.checkouts.status IS 'Order status: CREATED, PROCESSING, SHIPPED, DELIVERED, CANCELLED';


-- public.promotion_applied definition

-- Drop table

-- DROP TABLE promotion_applied;

CREATE TABLE promotion_applied (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	checkout_id uuid NOT NULL,
	promotion_id uuid NOT NULL,
	description text NOT NULL,
	discount numeric(10, 2) NOT NULL,
	CONSTRAINT promotion_applied_pkey PRIMARY KEY (id),
	CONSTRAINT promotion_applied_checkout_id_fkey FOREIGN KEY (checkout_id) REFERENCES checkouts(id),
	CONSTRAINT promotion_applied_checkout_id_fkey1 FOREIGN KEY (checkout_id) REFERENCES checkouts(id) ON DELETE CASCADE,
	CONSTRAINT promotion_applied_promotion_id_fkey FOREIGN KEY (promotion_id) REFERENCES promotions(id)
);


-- public.refresh_tokens definition

-- Drop table

-- DROP TABLE refresh_tokens;

CREATE TABLE refresh_tokens (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	user_id uuid NOT NULL,
	"token" varchar(255) NOT NULL,
	expires_at timestamptz NOT NULL,
	created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
	CONSTRAINT refresh_tokens_pkey PRIMARY KEY (id),
	CONSTRAINT refresh_tokens_token_key UNIQUE (token),
	CONSTRAINT refresh_tokens_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX idx_refresh_tokens_token ON public.refresh_tokens USING btree (token);
CREATE INDEX idx_refresh_tokens_user_id ON public.refresh_tokens USING btree (user_id);


-- public.checkout_items definition

-- Drop table

-- DROP TABLE checkout_items;

CREATE TABLE checkout_items (
	id uuid DEFAULT gen_random_uuid() NOT NULL,
	checkout_id uuid NOT NULL,
	product_id uuid NOT NULL,
	product_sku varchar(50) NOT NULL,
	product_name varchar(255) NOT NULL,
	quantity int4 NOT NULL,
	unit_price numeric(10, 2) NOT NULL,
	subtotal numeric(10, 2) NOT NULL,
	discount numeric(10, 2) NOT NULL,
	total numeric(10, 2) NOT NULL,
	CONSTRAINT checkout_items_pkey PRIMARY KEY (id),
	CONSTRAINT checkout_items_checkout_id_fkey FOREIGN KEY (checkout_id) REFERENCES checkouts(id),
	CONSTRAINT checkout_items_checkout_id_fkey1 FOREIGN KEY (checkout_id) REFERENCES checkouts(id) ON DELETE CASCADE,
	CONSTRAINT checkout_items_product_id_fkey FOREIGN KEY (product_id) REFERENCES products(id)
);


-- SEED DATA

INSERT INTO promotions
(id, "type", description, "rule", active, created_at, updated_at, deleted_at)
VALUES('9e66dd45-2b2e-4268-9537-ecdf68c6e3fc'::uuid, 'BUY_ONE_GET_ONE_FREE', 'Each MacBook Pro purchase comes with a free Raspberry Pi B', '{"free_sku": "234234", "trigger_sku": "43N23P", "free_quantity": 1, "trigger_quantity": 1}'::jsonb, true, '2025-05-17 18:40:03.192', '2025-05-17 18:40:03.192', NULL);
INSERT INTO promotions
(id, "type", description, "rule", active, created_at, updated_at, deleted_at)
VALUES('2569b8d5-5726-403e-aab2-be7ace3dd9e7'::uuid, 'BUY_3_PAY_2', 'When you buy 3 Google Home devices, you only pay for 2', '{"sku": "120P90", "min_quantity": 3, "free_quantity_divisor": 1, "paid_quantity_divisor": 2}'::jsonb, true, '2025-05-17 18:40:03.192', '2025-05-17 18:40:03.192', NULL);
INSERT INTO promotions
(id, "type", description, "rule", active, created_at, updated_at, deleted_at)
VALUES('a8bc46fe-731c-46e9-b8dd-aa7159d1ed70'::uuid, 'BULK_DISCOUNT', 'Buying more than 3 Alexa Speakers gives a 10% discount on all Alexa Speakers', '{"sku": "A304SD", "min_quantity": 4, "discount_percentage": 10}'::jsonb, true, '2025-05-17 18:40:03.192', '2025-05-17 18:40:03.192', NULL);


