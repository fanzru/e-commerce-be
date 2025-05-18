/**
 * E-Commerce Frontend JavaScript
 */

// API base URL
const API_BASE_URL = "/api/v1";

// Helper function for making API requests
async function fetchApi(endpoint, options = {}) {
  const token = localStorage.getItem("token");

  // Check if token is expired and try to refresh it
  if (token && auth.isTokenExpired()) {
    console.log("Token expired before request, attempting to refresh");
    // auth.logout();
    // throw new Error("Your session has expired. Please log in again.");
  }

  const defaultOptions = {
    headers: {
      "Content-Type": "application/json",
      ...(token && { Authorization: `Bearer ${token}` }),
    },
  };

  const requestOptions = {
    ...defaultOptions,
    ...options,
  };

  console.log(`API Request: ${endpoint}`, {
    method: requestOptions.method || "GET",
    headers: requestOptions.headers,
    hasBody: !!requestOptions.body,
  });

  try {
    const response = await fetch(`${API_BASE_URL}${endpoint}`, requestOptions);

    // Try to parse the response as JSON
    let data;
    const contentType = response.headers.get("content-type");
    if (contentType && contentType.includes("application/json")) {
      data = await response.json();
    } else {
      data = await response.text();
    }

    console.log(`API Response: ${endpoint}`, {
      status: response.status,
      ok: response.ok,
      contentType,
      data,
    });

    if (!response.ok) {
      const errorMessage =
        typeof data === "object" && data.message
          ? data.message
          : `API error: ${response.status}`;

      // Handle authentication errors
      if (
        response.status === 401 ||
        (typeof data === "object" &&
          data.code === "ERROR" &&
          (data.message === "Invalid token" || data.message.includes("token")))
      ) {
        console.error("Authentication error:", data);
        // auth.logout();
        throw new Error("Your session has expired. Please log in again.");
      }

      throw new Error(errorMessage);
    }

    return data;
  } catch (error) {
    console.error(`API Error: ${endpoint}`, error);
    throw error;
  }
}

// Authentication functions
const auth = {
  // Login user
  async login(email, password) {
    console.log("Login attempt for:", email);

    try {
      const response = await fetchApi("/auth/login", {
        method: "POST",
        body: JSON.stringify({ email, password }),
      });

      console.log("Login API response:", response);

      // Handle the API response format: { code, data: { access_token, refresh_token, expires_in }, message, server_time }
      if (response.data && response.data.access_token) {
        localStorage.setItem("token", response.data.access_token);
        if (response.data.refresh_token) {
          localStorage.setItem("refresh_token", response.data.refresh_token);
        }

        // Store token expiration time
        if (response.data.expires_in) {
          const expiresAt = Date.now() + response.data.expires_in * 1000;
          localStorage.setItem("token_expires_at", expiresAt.toString());
        }

        // If there's user data in the response, store it
        if (response.data.user) {
          localStorage.setItem("user", JSON.stringify(response.data.user));
        } else {
          // Create a minimal user object to ensure we have something in storage
          const userObj = { email: email };
          localStorage.setItem("user", JSON.stringify(userObj));
        }

        console.log(
          "Login successful, token saved:",
          response.data.access_token
        );
        return response.data;
      } else if (response.code === "SUCCESS" && response.data) {
        // Try alternate format where data might contain the token directly
        console.log("Trying alternate response format");

        if (response.data.access_token) {
          localStorage.setItem("token", response.data.access_token);
          localStorage.setItem("user", JSON.stringify({ email: email }));
          console.log("Login successful with alternate format");
          return response.data;
        }
      }

      console.error("Invalid login response format:", response);
      throw new Error("Invalid login response format");
    } catch (error) {
      console.error("Login error:", error);
      throw error;
    }
  },

  // Refresh token
  async refreshToken() {
    const refreshToken = localStorage.getItem("refresh_token");
    if (!refreshToken) {
      return false;
    }

    try {
      const response = await fetchApi("/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });

      if (response.data && response.data.access_token) {
        localStorage.setItem("token", response.data.access_token);

        // Update expiration time
        if (response.data.expires_in) {
          const expiresAt = Date.now() + response.data.expires_in * 1000;
          localStorage.setItem("token_expires_at", expiresAt.toString());
        }

        console.log("Token refreshed successfully");
        return true;
      }

      return false;
    } catch (error) {
      console.error("Token refresh failed:", error);
      return false;
    }
  },

  // Check if token has expired
  isTokenExpired() {
    const expiresAt = localStorage.getItem("token_expires_at");
    if (!expiresAt) return false;

    // Return true if token expiration time is in the past
    return Date.now() > parseInt(expiresAt, 10);
  },

  // Register new user
  async register(userData) {
    return fetchApi("/auth/register", {
      method: "POST",
      body: JSON.stringify(userData),
    });
  },

  // Logout user
  logout() {
    localStorage.removeItem("token");
    localStorage.removeItem("refresh_token");
    localStorage.removeItem("token_expires_at");
    localStorage.removeItem("user");
    window.location.href = "login.html";
  },

  // Get current user
  getCurrentUser() {
    const userJson = localStorage.getItem("user");
    return userJson ? JSON.parse(userJson) : null;
  },

  // Check if user is logged in
  isLoggedIn() {
    const hasToken = !!localStorage.getItem("token");

    // If we have a token but it's expired, try to refresh it
    if (hasToken && this.isTokenExpired()) {
      console.log("Token expired, attempting to refresh");
      // Note: we don't await this because we want a synchronous response,
      // the token will be refreshed in the background
      this.refreshToken().then((success) => {
        if (!success) {
          // If refresh fails, force logout
          //   this.logout();
        }
      });
    }

    return hasToken;
  },
};

// Product functions
const products = {
  // Get all products
  async getAll() {
    try {
      console.log("Fetching products from API...");
      const response = await fetchApi("/products");
      console.log("Raw API response for products:", response);
      return response;
    } catch (error) {
      console.error("Error fetching products:", error);
      throw error;
    }
  },

  // Get product by ID
  async getById(id) {
    try {
      return await fetchApi(`/products/${id}`);
    } catch (error) {
      console.error(`Error fetching product ${id}:`, error);
      throw error;
    }
  },
};

// Cart functions
const cart = {
  // Get user's cart
  async getCart() {
    if (!auth.isLoggedIn()) return null;

    try {
      console.log("Fetching user cart...");
      const response = await fetchApi("/carts/me");
      console.log("Raw API response for cart:", response);
      return response;
    } catch (error) {
      console.error("Error fetching cart:", error);

      // If unauthorized, force logout
      if (error.message.includes("session has expired")) {
        // auth.logout();
        return null;
      }

      // If cart not found (404), return null instead of throwing
      if (error.message.includes("404")) {
        console.log("No cart found for user, will create a new one");
        return null;
      }

      return null;
    }
  },

  // Create a new cart
  async createCart() {
    try {
      console.log("Creating new cart...");
      const response = await fetchApi("/carts", {
        method: "POST",
        body: JSON.stringify({}),
      });
      console.log("Cart creation response:", response);
      return response;
    } catch (error) {
      console.error("Error creating cart:", error);

      // If unauthorized, force logout
      if (error.message.includes("session has expired")) {
        // auth.logout();
      }

      throw error;
    }
  },

  // Add item to cart
  async addToCart(cartId, productId, quantity = 1) {
    try {
      console.log(`Adding product ${productId} to cart ${cartId}...`);
      const response = await fetchApi(`/carts/${cartId}/items`, {
        method: "POST",
        body: JSON.stringify({
          product_id: productId,
          quantity: quantity,
        }),
      });
      console.log("Add to cart response:", response);
      return response;
    } catch (error) {
      console.error(`Error adding product ${productId} to cart:`, error);
      throw error;
    }
  },

  // Add item to current user's cart
  async addToCurrentUserCart(productId, quantity = 1) {
    try {
      console.log(`Adding product ${productId} to current user's cart...`);
      const response = await fetchApi(`/carts/me`, {
        method: "POST",
        body: JSON.stringify({
          product_id: productId,
          quantity: quantity,
        }),
      });
      console.log("Add to current user's cart response:", response);
      return response;
    } catch (error) {
      console.error(
        `Error adding product ${productId} to current user's cart:`,
        error
      );
      throw error;
    }
  },

  // Remove item from cart
  async removeFromCart(cartId, itemId) {
    try {
      console.log(`Removing item ${itemId} from cart ${cartId}...`);
      const response = await fetchApi(`/carts/${cartId}/items/${itemId}`, {
        method: "DELETE",
      });
      console.log("Remove from cart response:", response);
      return response;
    } catch (error) {
      console.error(`Error removing item ${itemId} from cart:`, error);
      throw error;
    }
  },

  // Update cart item quantity
  async updateCartItem(cartId, itemId, quantity) {
    try {
      console.log(
        `Updating item ${itemId} in cart ${cartId} to quantity ${quantity}...`
      );
      const response = await fetchApi(`/carts/${cartId}/items/${itemId}`, {
        method: "PUT",
        body: JSON.stringify({
          quantity: quantity,
        }),
      });
      console.log("Update cart item response:", response);
      return response;
    } catch (error) {
      console.error(`Error updating item ${itemId} in cart:`, error);
      throw error;
    }
  },
};

// Form validation functions
const validateEmail = (email) => {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(String(email).toLowerCase());
};

// Initialize page-specific functionality
function initPage() {
  // Get the current page filename
  const path = window.location.pathname;
  const currentPage = path.substring(path.lastIndexOf("/") + 1) || "index.html";

  // Check if user is logged in
  if (
    !auth.isLoggedIn() &&
    !["login.html", "register.html", "index.html", ""].includes(currentPage)
  ) {
    window.location.href = "login.html";
    return;
  }

  // Update nav based on auth status
  updateNavigation();

  // Initialize page-specific JS
  switch (currentPage) {
    case "login.html":
      initLoginPage();
      break;
    case "register.html":
      initRegisterPage();
      break;
    case "products.html":
      initProductsPage();
      break;
    case "cart.html":
      initCartPage();
      break;
    case "index.html":
    case "":
      initHomePage();
      break;
  }
}

// Update navigation based on authentication status
function updateNavigation() {
  const navLinks = document.querySelector(".nav-links");
  if (!navLinks) return;

  if (auth.isLoggedIn()) {
    const user = auth.getCurrentUser();
    navLinks.innerHTML = `
      <a href="products.html">Products</a>
      <a href="cart.html">Cart</a>
      <a href="#" id="logout-link">Logout</a>
    `;

    // Add logout listener
    document.getElementById("logout-link").addEventListener("click", (e) => {
      e.preventDefault();
      //   auth.logout();
    });
  } else {
    navLinks.innerHTML = `
      <a href="login.html">Login</a>
      <a href="register.html">Register</a>
    `;
  }
}

// Initialize login page
function initLoginPage() {
  const loginForm = document.getElementById("login-form");
  if (!loginForm) return;

  // Add a status element for feedback
  const statusDiv = document.createElement("div");
  statusDiv.id = "login-status";
  statusDiv.style.marginTop = "10px";
  statusDiv.style.padding = "10px";
  statusDiv.style.borderRadius = "4px";
  statusDiv.style.display = "none";
  loginForm.appendChild(statusDiv);

  loginForm.addEventListener("submit", async (e) => {
    e.preventDefault();

    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    // Clear any previous status
    statusDiv.textContent = "";
    statusDiv.style.display = "none";
    statusDiv.className = "";

    // Show loading status
    statusDiv.textContent = "Logging in...";
    statusDiv.style.display = "block";
    statusDiv.style.backgroundColor = "#f8f9fa";
    statusDiv.style.color = "#333";

    try {
      const result = await auth.login(email, password);

      // Show success message
      statusDiv.textContent = "Login successful! Redirecting...";
      statusDiv.style.backgroundColor = "#d4edda";
      statusDiv.style.color = "#155724";

      console.log("Login successful, redirecting to products page");
      setTimeout(() => {
        window.location.href = "products.html";
      }, 1000); // Short delay to show success message
    } catch (error) {
      // Show error message
      console.error("Login error:", error);
      statusDiv.textContent =
        error.message || "Login failed. Please try again.";
      statusDiv.style.backgroundColor = "#f8d7da";
      statusDiv.style.color = "#721c24";
    }
  });
}

// Initialize register page
function initRegisterPage() {
  const registerForm = document.getElementById("register-form");
  if (!registerForm) return;

  registerForm.addEventListener("submit", async (e) => {
    e.preventDefault();

    const name = document.getElementById("name").value;
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    if (!validateEmail(email)) {
      alert("Please enter a valid email address");
      return;
    }

    try {
      await auth.register({ name, email, password });
      alert("Registration successful! Please log in.");
      window.location.href = "login.html";
    } catch (error) {
      alert(error.message || "Registration failed. Please try again.");
    }
  });
}

// Initialize products page
async function initProductsPage() {
  const productGrid = document.querySelector(".product-grid");
  if (!productGrid) return;

  // Show loading state
  productGrid.innerHTML =
    '<div class="loading-spinner">Loading products...</div>';

  try {
    const response = await products.getAll();
    console.log("Products API response:", response);

    // Handle different response structures
    let productList = [];

    // Check for response format: { code, data: { products: [...] }, message }
    if (
      response &&
      response.data &&
      response.data.products &&
      Array.isArray(response.data.products)
    ) {
      productList = response.data.products;
    }
    // Check for standard API response format: { code, data: [...], message }
    else if (response && response.data && Array.isArray(response.data)) {
      productList = response.data;
    }
    // Check if response itself is an array
    else if (Array.isArray(response)) {
      productList = response;
    }
    // Check if response has items property
    else if (response && response.items && Array.isArray(response.items)) {
      productList = response.items;
    }
    // Check if response has product_list property
    else if (
      response &&
      response.product_list &&
      Array.isArray(response.product_list)
    ) {
      productList = response.product_list;
    }
    // No valid products found
    else {
      console.error("Unexpected API response format:", response);
      productGrid.innerHTML =
        '<p class="text-center">No products available. Unexpected data format.</p>';
      return;
    }

    if (productList.length === 0) {
      productGrid.innerHTML =
        '<p class="text-center">No products available</p>';
      return;
    }

    // Log the product data structure to help debugging
    console.log("First product data structure:", productList[0]);

    productGrid.innerHTML = productList
      .map((product) => {
        // Extract product properties safely
        const name = product.name || product.product_name || "Unnamed Product";
        const price = product.price || product.unit_price || 0;
        const inventory =
          product.inventory_quantity || product.stock || product.inventory || 0;
        const id = product.id || product.product_id || "";
        const imageUrl =
          product.image_url || product.image || "img/product-placeholder.svg";

        return `
      <div class="product-card">
        <div class="product-image">
          <img src="${imageUrl}" alt="${name}">
        </div>
        <div class="product-info">
          <h3 class="product-name">${name}</h3>
          <p class="product-price">$${parseFloat(price).toFixed(2)}</p>
          <p class="product-inventory">In stock: ${inventory}</p>
          <button class="btn btn-block add-to-cart" data-product-id="${id}">Add to Cart</button>
        </div>
      </div>
      `;
      })
      .join("");

    // Add event listeners for add to cart buttons
    document.querySelectorAll(".add-to-cart").forEach((button) => {
      button.addEventListener("click", async () => {
        if (!auth.isLoggedIn()) {
          window.location.href = "login.html";
          return;
        }

        const productId = button.dataset.productId;
        const originalText = button.textContent;

        // Show loading state
        button.textContent = "Adding...";
        button.disabled = true;

        try {
          // First try to add to current user's cart directly
          try {
            await cart.addToCurrentUserCart(productId, 1);
            button.textContent = "Added!";
            setTimeout(() => {
              button.textContent = originalText;
              button.disabled = false;
            }, 2000);
            return;
          } catch (directAddError) {
            console.log(
              "Could not add directly to user cart, trying fallback method...",
              directAddError
            );
          }

          // Fallback: Get or create cart
          let userCart = await cart.getCart();
          if (!userCart) {
            userCart = await cart.createCart();
          }

          if (!userCart || !userCart.id) {
            throw new Error(
              "Could not access or create your cart. Please try logging in again."
            );
          }

          await cart.addToCart(userCart.id, productId, 1);
          button.textContent = "Added!";
          setTimeout(() => {
            button.textContent = originalText;
            button.disabled = false;
          }, 2000);
        } catch (error) {
          console.error("Add to cart error:", error);
          button.textContent = "Failed";
          alert(error.message || "Failed to add product to cart");

          setTimeout(() => {
            button.textContent = originalText;
            button.disabled = false;
          }, 2000);
        }
      });
    });
  } catch (error) {
    console.error("Error loading products:", error);
    productGrid.innerHTML = `<p class="text-center">Error loading products: ${error.message}</p>`;
  }
}

// Initialize cart page
async function initCartPage() {
  const cartContainer = document.querySelector(".cart-container");
  if (!cartContainer) return;

  // Show loading state
  cartContainer.innerHTML =
    '<div class="loading-spinner">Loading your cart...</div>';

  try {
    const response = await cart.getCart();
    console.log("Cart API response:", response);

    // Handle different response structures
    let userCart = null;

    // Check for standard API response format: { code, data: {...}, message }
    if (response && response.data) {
      userCart = response.data;
    }
    // Check if response itself is a cart object
    else if (response && response.id) {
      userCart = response;
    }

    // If no cart exists yet, show empty cart message
    if (!userCart) {
      cartContainer.innerHTML = `
        <div class="text-center">
          <p>Your cart is empty</p>
          <a href="products.html" class="btn mt-3">Browse Products</a>
        </div>
      `;
      return;
    }

    // Extract cart items safely
    let cartItems = [];
    if (userCart && userCart.items && Array.isArray(userCart.items)) {
      cartItems = userCart.items;
    } else if (
      userCart &&
      userCart.cart_items &&
      Array.isArray(userCart.cart_items)
    ) {
      cartItems = userCart.cart_items;
    }

    if (cartItems.length === 0) {
      cartContainer.innerHTML = `
        <div class="text-center">
          <p>Your cart is empty</p>
          <a href="products.html" class="btn mt-3">Browse Products</a>
        </div>
      `;
      return;
    }

    let total = 0;

    const itemsHtml = cartItems
      .map((item) => {
        // Extract item properties safely
        const product = item.product || {};
        const name = product.name || item.product_name || "Unnamed Product";
        const price = product.price || item.unit_price || 0;
        const quantity = item.quantity || 1;
        const itemId = item.id || item.item_id || "";
        const maxQuantity = product.inventory || 10; // Default to 10 if inventory unknown

        const itemTotal = quantity * price;
        total += itemTotal;

        return `
        <div class="cart-item" data-item-id="${itemId}">
          <div class="cart-item-info">
            <h3>${name}</h3>
            <p>Price: $${parseFloat(price).toFixed(2)}</p>
            <div class="quantity-controls">
              <span>Quantity: </span>
              <button class="quantity-btn decrease" ${
                quantity <= 1 ? "disabled" : ""
              }>-</button>
              <span class="quantity-value">${quantity}</span>
              <button class="quantity-btn increase" ${
                quantity >= maxQuantity ? "disabled" : ""
              }>+</button>
            </div>
            <p>Subtotal: $<span class="item-subtotal">${itemTotal.toFixed(
              2
            )}</span></p>
          </div>
          <div>
            <button class="btn remove-item">Remove</button>
          </div>
        </div>
      `;
      })
      .join("");

    cartContainer.innerHTML = `
      <div class="cart-items">
        ${itemsHtml}
      </div>
      <div class="cart-summary">
        <h3>Order Summary</h3>
        <p class="cart-total">Total: $<span id="cart-total-value">${total.toFixed(
          2
        )}</span></p>
        <button class="btn btn-block checkout-btn">Proceed to Checkout</button>
      </div>
    `;

    // Add event listeners for quantity controls
    document.querySelectorAll(".cart-item").forEach((cartItem) => {
      const itemId = cartItem.dataset.itemId;
      const decreaseBtn = cartItem.querySelector(".decrease");
      const increaseBtn = cartItem.querySelector(".increase");
      const quantityEl = cartItem.querySelector(".quantity-value");
      const subtotalEl = cartItem.querySelector(".item-subtotal");

      // Find the corresponding cart item in our data
      const itemData = cartItems.find(
        (item) => item.id === itemId || item.item_id === itemId
      );
      if (!itemData) return;

      const price =
        itemData.unit_price ||
        (itemData.product && itemData.product.price) ||
        0;

      // Handle decrease button
      decreaseBtn.addEventListener("click", async () => {
        let currentQty = parseInt(quantityEl.textContent);
        if (currentQty <= 1) return;

        const newQty = currentQty - 1;
        decreaseBtn.disabled = true;

        try {
          await cart.updateCartItem(userCart.id, itemId, newQty);

          // Update UI
          quantityEl.textContent = newQty;
          decreaseBtn.disabled = newQty <= 1;
          increaseBtn.disabled = false;

          // Update subtotal
          const newSubtotal = (newQty * price).toFixed(2);
          subtotalEl.textContent = newSubtotal;

          // Update cart total
          updateCartTotal();
        } catch (error) {
          alert("Failed to update quantity");
          decreaseBtn.disabled = false;
        }
      });

      // Handle increase button
      increaseBtn.addEventListener("click", async () => {
        let currentQty = parseInt(quantityEl.textContent);
        const maxQty = itemData.product?.inventory || 10;

        if (currentQty >= maxQty) return;

        const newQty = currentQty + 1;
        increaseBtn.disabled = true;

        try {
          await cart.updateCartItem(userCart.id, itemId, newQty);

          // Update UI
          quantityEl.textContent = newQty;
          increaseBtn.disabled = newQty >= maxQty;
          decreaseBtn.disabled = false;

          // Update subtotal
          const newSubtotal = (newQty * price).toFixed(2);
          subtotalEl.textContent = newSubtotal;

          // Update cart total
          updateCartTotal();
        } catch (error) {
          alert("Failed to update quantity");
          increaseBtn.disabled = false;
        }
      });
    });

    // Function to update the cart total
    function updateCartTotal() {
      let newTotal = 0;
      document.querySelectorAll(".cart-item").forEach((item) => {
        const quantity = parseInt(
          item.querySelector(".quantity-value").textContent
        );
        const itemSubtotal = parseFloat(
          item.querySelector(".item-subtotal").textContent
        );
        newTotal += itemSubtotal;
      });
      document.getElementById("cart-total-value").textContent =
        newTotal.toFixed(2);
    }

    // Add event listeners for remove buttons
    document.querySelectorAll(".remove-item").forEach((button) => {
      button.addEventListener("click", async () => {
        const itemId = button.closest(".cart-item").dataset.itemId;
        const cartId = userCart.id;

        try {
          await cart.removeFromCart(cartId, itemId);
          // Refresh page after removal
          window.location.reload();
        } catch (error) {
          alert(error.message || "Failed to remove item from cart");
        }
      });
    });

    // Add checkout button listener
    document.querySelector(".checkout-btn").addEventListener("click", () => {
      // Create a checkout using the current cart
      createCheckout(userCart.id);
    });

    // Function to create a checkout
    async function createCheckout(cartId) {
      try {
        const response = await fetch(`${API_BASE_URL}/checkouts`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${localStorage.getItem("token")}`,
          },
          body: JSON.stringify({
            cart_id: cartId,
          }),
        });

        const data = await response.json();

        if (response.ok) {
          alert("Checkout successful! Order ID: " + data.data.id);
          // Redirect to a confirmation page or reload
          window.location.reload();
        } else {
          alert("Checkout failed: " + (data.message || "Unknown error"));
        }
      } catch (error) {
        alert("Checkout failed: " + error.message);
      }
    }
  } catch (error) {
    console.error("Error loading cart:", error);
    cartContainer.innerHTML = `<p class="text-center">Error loading cart: ${error.message}</p>`;
  }
}

// Initialize home page
function initHomePage() {
  // Redirect to products if logged in
  if (auth.isLoggedIn()) {
    window.location.href = "products.html";
  }
}

// Run initialization when DOM is loaded
document.addEventListener("DOMContentLoaded", initPage);
