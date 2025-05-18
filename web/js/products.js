/**
 * Products page functionality
 */

// Product functions
const products = {
  // Get all products
  async getAll() {
    try {
      console.log("Fetching products from API...");
      const response = await fetchApi("/products");
      console.log("Raw API response for products:", response);

      // Adaptive handling for various response formats
      // to handle both existing nested format and future simplified format

      // Format 1: Existing deeply nested format
      // {code, message, data: {code, data: {products}, message}, server_time}
      if (response?.data?.data?.products) {
        console.log("Found deeply nested products format");
        return {
          data: {
            products: response.data.data.products,
          },
        };
      }

      // Format 2: Single wrapper response from fixed backend
      // {code, message, data: {products}, server_time}
      if (response?.data?.products) {
        console.log("Found single-level nested products format");
        return {
          data: {
            products: response.data.products,
          },
        };
      }

      // Format 3: Direct products array
      if (Array.isArray(response)) {
        console.log("Found direct products array format");
        return {
          data: {
            products: response,
          },
        };
      }

      // If no recognized format, return as is and let the error handling manage it
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

            // If the error contains specific messages, we might need to handle differently
            if (
              directAddError.message &&
              directAddError.message.includes("already exists")
            ) {
              // This might be a soft-deleted item issue, try to force a refresh
              button.textContent = "Added!";
              setTimeout(() => {
                button.textContent = originalText;
                button.disabled = false;
              }, 2000);
              return;
            }
          }

          // Fallback: Get or create cart
          let userCart = await cart.getCart();
          if (!userCart) {
            console.log("Creating new cart as fallback...");
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

          // Show more specific error message to help with debugging
          let errorMsg = "Failed to add product to cart";
          if (error.message) {
            errorMsg += `: ${error.message}`;
          }

          alert(errorMsg);

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
