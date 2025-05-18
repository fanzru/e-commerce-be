/**
 * Common JavaScript for E-Commerce Frontend
 * Contains shared functionality used across all pages
 */

// Helper function to format price in USD format
function formatPrice(price) {
  return "USD " + parseFloat(price).toFixed(2);
}

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
    console.log("Logging out user");
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
          // this.logout();
        }
      });
    }

    return hasToken;
  },
};

// Update navigation based on authentication status
function updateNavigation() {
  // Get all navigation containers across different pages
  const navElements = document.querySelectorAll("header nav .flex");
  if (navElements.length === 0) return;

  const currentPath = window.location.pathname;
  const isProductsPage = currentPath.includes("products.html");
  const isLoginPage = currentPath.includes("login.html");
  const isRegisterPage = currentPath.includes("register.html");
  const isCartPage = currentPath.includes("cart.html");

  if (auth.isLoggedIn()) {
    // User is logged in - show Products, Cart, Logout
    navElements.forEach(function (nav) {
      nav.innerHTML =
        '<a href="products.html" class="text-gray-600 hover:text-gray-900 ' +
        (isProductsPage ? "text-gray-900 font-medium" : "") +
        '">Products</a>' +
        '<a href="cart.html" class="text-gray-600 hover:text-gray-900 ' +
        (isCartPage ? "text-gray-900 font-medium" : "") +
        '">Cart</a>' +
        '<a href="#" id="logout-link" class="text-gray-600 hover:text-gray-900">Logout</a>';
    });

    // Add logout listeners to all logout links
    var logoutLinks = document.querySelectorAll("#logout-link");
    for (var i = 0; i < logoutLinks.length; i++) {
      logoutLinks[i].addEventListener("click", function (e) {
        e.preventDefault();
        console.log("Logout clicked");
        auth.logout();
      });
    }
  } else {
    // User is not logged in - show Products, Login, Register
    navElements.forEach(function (nav) {
      nav.innerHTML =
        '<a href="products.html" class="text-gray-600 hover:text-gray-900 ' +
        (isProductsPage ? "text-gray-900 font-medium" : "") +
        '">Products</a>' +
        '<a href="login.html" class="text-gray-600 hover:text-gray-900 ' +
        (isLoginPage ? "text-gray-900 font-medium" : "") +
        '">Login</a>' +
        '<a href="register.html" class="text-gray-600 hover:text-gray-900 ' +
        (isRegisterPage ? "text-gray-900 font-medium" : "") +
        '">Register</a>';
    });
  }
}

// Form validation functions
const validateEmail = (email) => {
  const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  return re.test(String(email).toLowerCase());
};

// Initialize page-specific functionality
function initPage() {
  console.log("Initializing page");

  // Check if user is logged in first
  const isLoggedIn = auth.isLoggedIn();
  console.log("Auth status:", isLoggedIn ? "Logged in" : "Not logged in");

  // Get current path
  const path = window.location.pathname;

  // Immediately handle auth-based redirections before any other initialization
  if (isLoggedIn) {
    // If logged in and on login or register page, redirect to products immediately
    if (path.includes("login.html") || path.includes("register.html")) {
      console.log(
        "User already logged in, redirecting from auth page to products"
      );
      window.location.href = "products.html";
      return; // Exit early to prevent further execution
    }
  } else {
    // If not logged in and on protected pages, redirect to login
    if (
      path.includes("cart.html") ||
      path.includes("checkout.html") ||
      (!path.includes("login.html") &&
        !path.includes("register.html") &&
        !path.includes("index.html") &&
        !path.includes("products.html") &&
        path !== "/" &&
        path !== "")
    ) {
      console.log("User not logged in, redirecting to login");
      window.location.href = "login.html";
      return; // Exit early
    }
  }

  // Update navigation links
  updateNavigation();

  // Initialize page-specific functions based on the current page
  if (path.includes("login.html")) {
    if (typeof initLoginPage === "function") {
      initLoginPage();
    }
  } else if (path.includes("register.html")) {
    if (typeof initRegisterPage === "function") {
      initRegisterPage();
    }
  } else if (path.includes("products.html")) {
    if (typeof initProductsPage === "function") {
      initProductsPage();
    }
  } else if (path.includes("cart.html")) {
    if (isLoggedIn && typeof initCartPage === "function") {
      initCartPage();
    }
  } else if (path.includes("checkout.html")) {
    if (isLoggedIn && typeof initCheckoutPage === "function") {
      initCheckoutPage();
    }
  } else if (path.includes("index.html") || path === "/" || path === "") {
    if (typeof initHomePage === "function") {
      initHomePage();
    }
  }
}

// Run initialization when DOM is loaded
document.addEventListener("DOMContentLoaded", initPage);
