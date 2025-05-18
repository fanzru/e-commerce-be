# FORTEPAY E-Commerce Web UI

This folder contains a simple client-side web UI for the FORTEPAY E-Commerce backend. It allows customers to:

- Register for a new account
- Login to their account
- Browse products
- Add items to cart
- View and manage their shopping cart

## Structure

- `index.html` - Landing page
- `login.html` - User login page
- `register.html` - New user registration page
- `products.html` - Product listing page
- `cart.html` - Shopping cart page

## Assets

- `css/` - Stylesheets
- `js/` - JavaScript files
- `img/` - Images and icons

## Features

The UI implements the following e-commerce functionality:

1. User authentication (login/register)
2. Product browsing
3. Cart management
4. Special promotions handling:
   - Each sale of a MacBook Pro comes with a free Raspberry Pi
   - Buy 3 Google Homes for the price of 2
   - 10% discount when buying more than 3 Alexa Speakers

## Usage

To use this UI, run the backend API server and then navigate to:

```
http://localhost:8080/web/
```

## Authentication

The UI uses JWT token authentication. After login, the token is stored in localStorage and automatically included in API requests.

## Development

This is a static client-side application that communicates with the backend API. To modify:

1. Edit HTML files to change the structure
2. Modify CSS in `css/style.css` to change the appearance
3. Update JavaScript in `js/main.js` to change the behavior

No build step is required as this is plain HTML/CSS/JavaScript.
