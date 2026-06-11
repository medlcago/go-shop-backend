# Go Shop Backend

## API Reference

### Users

| Method | Endpoint                                   | Description                         |
|--------|--------------------------------------------|-------------------------------------|
| POST   | `/api/v1/users/register`                   | Register                            |
| POST   | `/api/v1/users/login`                      | Login                               |
| GET    | `/api/v1/users/me`                         | Get current user (auth)             |
| POST   | `/api/v1/users/me/setup-2fa`               | Setup 2FA (auth)                    |
| POST   | `/api/v1/users/me/confirm-2fa`             | Confirm 2FA (auth)                  |
| POST   | `/api/v1/users/me/disable-2fa`             | Disable 2FA (auth)                  |
| POST   | `/api/v1/users/me/send-email-confirmation` | Send email confirmation code (auth) |
| POST   | `/api/v1/users/me/confirm-email`           | Confirm email address (auth)        |

### Products

| Method | Endpoint                                 | Description                 |
|--------|------------------------------------------|-----------------------------|
| GET    | `/api/v1/products`                       | List products               |
| GET    | `/api/v1/products/:id`                   | Get product                 |
| GET    | `/api/v1/products/search`                | Search products             |
| POST   | `/api/v1/products`                       | Create product (auth)       |
| PATCH  | `/api/v1/products/:id`                   | Update product (auth)       |
| POST   | `/api/v1/products/:id/images/upload-url` | Upload product image (auth) |
| POST   | `/api/v1/products/:id/images`            | Confirm image upload (auth) |

### Orders

| Method | Endpoint                           | Description           |
|--------|------------------------------------|-----------------------|
| GET    | `/api/v1/orders`                   | List my orders        |
| GET    | `/api/v1/orders/:id`               | Get order details     |
| POST   | `/api/v1/orders`                   | Create order (auth)   |
| POST   | `/api/v1/orders/:id/items`         | Add item              |
| DELETE | `/api/v1/orders/:id/items/:itemID` | Remove item           |
| DELETE | `/api/v1/orders/:id/items`         | Clear Items           |
| POST   | `/api/v1/orders/:id/checkout`      | Checkout order (auth) |
| POST   | `/api/v1/orders/:id/cancel`        | Cancel order (auth)   |

### Categories

| Method | Endpoint                 | Description        |
|--------|--------------------------|--------------------|
| GET    | `/api/v1/categories`     | List categories    |
| GET    | `/api/v1/categories/:id` | List subcategories |

### Payments

| Method | Endpoint                            | Description                                       |
|--------|-------------------------------------|---------------------------------------------------|
| POST   | `/api/v1/payments`                  | Create payment (auth)                             |
| GET    | `/api/v1/payments/webhook/yookassa` | Yookassa Webhook (yookassa IP whitelist, no auth) |

### Wishlists

| Method | Endpoint                                         | Description                      |
|--------|--------------------------------------------------|----------------------------------|
| GET    | `/api/v1/wishlists`                              | My wishlists (auth)              |
| GET    | `/api/v1/wishlists/:wishlistID`                  | Get wishlist details (auth)      |
| POST   | `/api/v1/wishlists`                              | Create wishlist (auth)           |
| PATH   | `/api/v1/wishlists/:wishlistID`                  | Update wishlist (auth)           |
| POST   | `/api/v1/wishlists/:wishlistID/regenerate-token` | Regenerate share token (auth)    |
| POST   | `/api/v1/wishlists/:wishlistID/items`            | Add item to wishlist (auth)      |
| PATCH  | `/api/v1/wishlists/:wishlistID/items/:itemID`    | Update item in wishlist (auth)   |
| DELETE | `/api/v1/wishlists/:wishlistID/items/:itemID`    | Remove item from wishlist (auth) |
| GET    | `/api/v1/wishlists/shared/:token`                | Get shared wishlist by token     |