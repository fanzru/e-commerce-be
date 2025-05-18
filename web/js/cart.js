/**
 * Shopping cart functionality
 */

// Cart functions
const cart = {
  // Get user's cart
  async getCart() {
    if (!auth.isLoggedIn()) return null;

    try {
      console.log("Fetching user cart...");
      const response = await fetchApi("/carts/me");
      console.log("Raw API response for cart:", response);

      // Handle different possible response formats

      // Format 1: Double nested response format (current)
      // {code, message, data: {code, data: {cart}, message}, server_time}
      if (response?.data?.data) {
        console.log("Found double-nested cart data format");
        return { data: response.data.data };
      }

      // Format 2: Single wrapper response (after backend fix)
      // {code, message, data: {cart}, server_time}
      if (response?.data) {
        console.log("Found single-level nested cart data format");
        return { data: response.data };
      }

      // Format 3: Direct cart object
      if (response && response.id) {
        console.log("Found direct cart object format");
        return { data: response };
      }

      return response;
    } catch (error) {
      console.error("Error fetching cart:", error);

      // If unauthorized, force logout
      if (error.message && error.message.includes("session has expired")) {
        // auth.logout();
        return null;
      }

      // If cart not found (404), return null instead of throwing
      if (error.message && error.message.includes("404")) {
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
      // First check if the product has enough stock
      try {
        const productResponse = await fetchApi(`/products/${productId}`);
        console.log("Product info response:", productResponse);

        // Check if we have product data
        let productData = null;
        if (productResponse.data) {
          productData = productResponse.data;
        } else if (productResponse.id) {
          productData = productResponse;
        }

        // Validate stock
        if (productData) {
          const availableStock =
            productData.stock || productData.inventory || 0;
          console.log(
            `Product ${productId} has ${availableStock} stock available`
          );

          if (availableStock < quantity) {
            throw new Error(
              `Not enough inventory. Available: ${availableStock}, Requested: ${quantity}`
            );
          }
        }
      } catch (productError) {
        console.error("Error checking product stock:", productError);
        // Only throw if it's a stock error, otherwise continue
        if (
          productError.message &&
          productError.message.includes("enough inventory")
        ) {
          throw productError;
        }
      }

      console.log(`Adding product ${productId} to current user's cart...`);

      // Use the correct endpoint: /carts/me
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

      // Check if it's a known error that we can recover from
      if (
        error.message &&
        (error.message.includes("cart item") ||
          error.message.includes("already exists") ||
          error.message.includes("deleted_at"))
      ) {
        console.warn(
          "Detected possible soft delete issue, trying fallback approach..."
        );

        try {
          // Get current cart to get its ID
          const cartResponse = await fetchApi("/carts/me");
          const cartId = cartResponse.data.id;

          // Try direct endpoint instead
          console.log(`Trying direct endpoint for cart ${cartId}...`);
          const directResponse = await fetchApi(`/carts/${cartId}/items`, {
            method: "POST",
            body: JSON.stringify({
              product_id: productId,
              quantity: quantity,
            }),
          });

          console.log("Direct endpoint response:", directResponse);
          return directResponse;
        } catch (fallbackError) {
          console.error("Fallback approach failed:", fallbackError);
          throw fallbackError;
        }
      }

      // Check if user is unauthorized
      if (
        error.message &&
        (error.message.includes("unauthorized") ||
          error.message.includes("Unauthorized") ||
          error.message.includes("session has expired") ||
          error.message.includes("token"))
      ) {
        console.error("Authentication error detected when adding to cart");
        // Redirect to login page
        window.location.href = "login.html";
        throw new Error("Please login to add items to your cart");
      }

      throw error;
    }
  },

  // Remove item from cart
  async removeFromCart(cartId, itemId) {
    try {
      console.log(`Removing item ${itemId} from cart...`);
      // Update to use the /carts/me/items/{itemId} endpoint
      const response = await fetchApi(`/carts/me/items/${itemId}`, {
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
      console.log(`Updating item ${itemId} to quantity ${quantity}...`);
      // Update to use the /carts/me/items/{itemId} endpoint
      const response = await fetchApi(`/carts/me/items/${itemId}`, {
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

// Initialize cart page
async function initCartPage() {
  const cartItemsContainer = document.getElementById("cart-items");
  if (!cartItemsContainer) return;

  // Show loading state
  cartItemsContainer.innerHTML =
    '<div class="p-6 text-center text-gray-500"><i class="fas fa-spinner fa-spin mr-2"></i>Loading your cart...</div>';

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
      cartItemsContainer.innerHTML = `
        <div class="p-8 text-center">
          <div class="text-gray-500 mb-4"><i class="fas fa-shopping-cart text-4xl mb-4"></i></div>
          <p class="text-lg text-gray-700 mb-4">Keranjang anda masih kosong</p>
          <a href="products.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
            <i class="fas fa-shopping-bag mr-2"></i> Lihat Produk
          </a>
        </div>
      `;
      updateCartSummary(0, 0, 0);
      return;
    }

    // Extract cart items safely
    let cartItems = [];
    if (userCart.items && Array.isArray(userCart.items)) {
      cartItems = userCart.items;
    } else if (userCart.cart_items && Array.isArray(userCart.cart_items)) {
      cartItems = userCart.cart_items;
    }

    if (cartItems.length === 0) {
      cartItemsContainer.innerHTML = `
        <div class="p-8 text-center">
          <div class="text-gray-500 mb-4"><i class="fas fa-shopping-cart text-4xl mb-4"></i></div>
          <p class="text-lg text-gray-700 mb-4">Keranjang anda masih kosong</p>
          <a href="products.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
            <i class="fas fa-shopping-bag mr-2"></i> Lihat Produk
          </a>
        </div>
      `;
      updateCartSummary(0, 0, 0);
      return;
    }

    let subtotal = 0;

    // Check if we have a custom template function defined in cart.html
    if (
      window.cartItemTemplate &&
      typeof window.cartItemTemplate === "function"
    ) {
      const itemsHtml = cartItems
        .map((item) => {
          // Extract item properties safely
          const product = item.product || {};
          const name = product.name || item.product_name || "Unnamed Product";
          const price = product.price || item.unit_price || 0;
          const quantity = item.quantity || 1;
          const itemId = item.id || item.item_id || "";
          const maxQuantity = product.inventory || 10; // Default to 10 if inventory unknown
          const sku = product.sku || item.product_sku || "";
          const imageUrl =
            product.image_url || product.image || "img/product-placeholder.svg";

          const itemTotal = quantity * price;
          subtotal += itemTotal;

          // Use the custom template
          return window.cartItemTemplate({
            id: itemId,
            quantity: quantity,
            product: {
              id: product.id || "",
              name: name,
              price: price,
              sku: sku,
              image: imageUrl,
              inventory: maxQuantity,
            },
          });
        })
        .join("");

      // Insert cart items HTML
      cartItemsContainer.innerHTML = itemsHtml;
    } else {
      // Use the fallback template
      const itemsHtml = cartItems
        .map((item) => {
          // Extract item properties safely
          const product = item.product || {};
          const name = product.name || item.product_name || "Unnamed Product";
          const price = product.price || item.unit_price || 0;
          const quantity = item.quantity || 1;
          const itemId = item.id || item.item_id || "";
          const maxQuantity = product.inventory || 10; // Default to 10 if inventory unknown
          const sku = product.sku || item.product_sku || "";
          const imageUrl =
            product.image_url || product.image || "img/product-placeholder.svg";

          const itemTotal = quantity * price;
          subtotal += itemTotal;

          return `
          <div class="cart-item p-4 border-b border-gray-200" data-item-id="${itemId}">
            <div class="flex flex-col sm:flex-row">
              <div class="flex-shrink-0 w-full sm:w-24 h-24 bg-gray-100 rounded-md overflow-hidden mb-4 sm:mb-0">
                <img src="${imageUrl}" alt="${name}" class="w-full h-full object-contain p-2">
              </div>
              <div class="flex-grow sm:ml-4">
                <div class="flex flex-col sm:flex-row justify-between">
                  <div>
                    <h3 class="text-base font-medium text-gray-800">${name}</h3>
                    ${
                      sku
                        ? `<p class="text-sm text-gray-500">SKU: ${sku}</p>`
                        : ""
                    }
                    <p class="text-base font-medium text-green-600 mt-1">${formatPrice(
                      price
                    )}</p>
                  </div>
                  <div class="flex items-center mt-4 sm:mt-0">
                    <div class="flex border border-gray-300 rounded-md">
                      <button class="decrease-quantity-btn px-3 py-1 text-gray-600 hover:bg-gray-100" data-item-id="${itemId}" ${
            quantity <= 1 ? "disabled" : ""
          }>
                        <i class="fas fa-minus"></i>
                      </button>
                      <input type="number" min="1" value="${quantity}" class="quantity-input w-12 text-center border-x border-gray-300" data-item-id="${itemId}">
                      <button class="increase-quantity-btn px-3 py-1 text-gray-600 hover:bg-gray-100" data-item-id="${itemId}" ${
            quantity >= maxQuantity ? "disabled" : ""
          }>
                        <i class="fas fa-plus"></i>
                      </button>
                    </div>
                    <button class="remove-item-btn ml-3 text-red-500 hover:text-red-700" data-item-id="${itemId}">
                      <i class="fas fa-trash-alt"></i>
                    </button>
                  </div>
                </div>
                <div class="mt-3">
                  <p class="text-gray-600 text-sm">Subtotal: <span class="font-medium">${formatPrice(
                    itemTotal
                  )}</span></p>
                </div>
              </div>
            </div>
          </div>
        `;
        })
        .join("");

      // Insert cart items HTML
      cartItemsContainer.innerHTML = itemsHtml;
    }

    // Check for promotions and display them
    const hasPromotions =
      userCart.applicable_promotions &&
      userCart.applicable_promotions.length > 0;
    const discount = hasPromotions
      ? parseFloat(userCart.potential_discount || 0)
      : 0;
    const finalTotal = subtotal - discount;

    // Update the cart summary
    updateCartSummary(subtotal, discount, finalTotal);

    // Show promotions if available
    if (hasPromotions) {
      const discountSection = document.getElementById("discount-section");
      const discountItems = document.getElementById("discount-items");

      if (discountSection) {
        discountSection.classList.remove("hidden");
      }

      if (discountItems) {
        const promotionsHtml = userCart.applicable_promotions
          .map((promo) => {
            const promoDiscount = parseFloat(promo.discount || 0);
            return `
              <div class="flex justify-between text-sm">
                <span>${promo.description || "Diskon"}</span>
                <span class="font-medium">-${formatPrice(promoDiscount)}</span>
              </div>
            `;
          })
          .join("");

        discountItems.innerHTML = promotionsHtml;
      }
    }

    // Add event listeners for quantity controls
    setupCartItemEventListeners(cartItems, userCart);

    // Setup checkout button
    const checkoutButton = document.getElementById("checkout-button");
    if (checkoutButton) {
      checkoutButton.addEventListener("click", () => {
        window.location.href = "checkout.html";
      });
    }
  } catch (error) {
    console.error("Error loading cart:", error);
    cartItemsContainer.innerHTML = `
      <div class="p-8 text-center">
        <div class="text-red-500 mb-4"><i class="fas fa-exclamation-circle text-4xl mb-4"></i></div>
        <p class="text-lg text-gray-700 mb-4">Error loading your cart: ${error.message}</p>
        <a href="products.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
          <i class="fas fa-shopping-bag mr-2"></i> View Products
        </a>
      </div>
    `;
  }
}

// Function to update cart total
function updateCartTotal() {
  let newTotal = 0;
  document.querySelectorAll(".cart-item").forEach((item) => {
    const subtotalText = item.querySelector(".mt-3 .font-medium").textContent;

    // Parse the formatted price back to a number
    const subtotal = parseFloat(subtotalText.replace("USD", "").trim());

    newTotal += subtotal;
  });

  // Refresh cart to get updated promotions and discounts
  refreshCartPromotions(newTotal);
}

// Function to refresh cart promotions
async function refreshCartPromotions(subtotal) {
  try {
    // Get updated cart with promotions
    const cartResponse = await cart.getCart();

    if (cartResponse && cartResponse.data) {
      const userCart = cartResponse.data;

      // Check for promotions and update them
      const hasPromotions =
        userCart.applicable_promotions &&
        userCart.applicable_promotions.length > 0;
      const discount = hasPromotions
        ? parseFloat(userCart.potential_discount || 0)
        : 0;
      const finalTotal = subtotal - discount;

      // Update the cart summary with fresh data
      updateCartSummary(subtotal, discount, finalTotal);

      // Show promotions if available
      if (hasPromotions) {
        const discountSection = document.getElementById("discount-section");
        const discountItems = document.getElementById("discount-items");

        if (discountSection) {
          discountSection.classList.remove("hidden");
        }

        if (discountItems) {
          const promotionsHtml = userCart.applicable_promotions
            .map((promo) => {
              const promoDiscount = parseFloat(promo.discount || 0);
              return `
                <div class="flex justify-between text-sm">
                  <span>${promo.description || "Diskon"}</span>
                  <span class="font-medium">-${formatPrice(
                    promoDiscount
                  )}</span>
                </div>
              `;
            })
            .join("");

          discountItems.innerHTML = promotionsHtml;
        }
      } else {
        // Hide discount section if no promotions
        const discountSection = document.getElementById("discount-section");
        if (discountSection) {
          discountSection.classList.add("hidden");
        }
      }
    } else {
      // If cart data is not available, just update with no discount
      updateCartSummary(subtotal, 0, subtotal);
    }
  } catch (error) {
    console.error("Error refreshing cart promotions:", error);
    // Fallback to updating summary without discounts
    updateCartSummary(subtotal, 0, subtotal);
  }
}

// Function to update cart summary
function updateCartSummary(subtotal, discount, total) {
  const subtotalEl = document.getElementById("cart-subtotal");
  const discountEl = document.getElementById("cart-discount");
  const totalEl = document.getElementById("cart-final-total");

  if (subtotalEl) subtotalEl.textContent = formatPrice(subtotal);
  if (discountEl) discountEl.textContent = `-${formatPrice(discount)}`;
  if (totalEl) totalEl.textContent = formatPrice(total);
}

// Setup cart item event listeners
function setupCartItemEventListeners(cartItems, userCart) {
  // Handle quantity inputs
  document.querySelectorAll(".quantity-input").forEach((input) => {
    const itemId = input.dataset.itemId;

    // Find the corresponding cart item in our data
    const itemData = cartItems.find(
      (item) => item.id === itemId || item.item_id === itemId
    );
    if (!itemData) return;

    const price = itemData.product?.price || itemData.unit_price || 0;
    const maxQuantity = itemData.product?.inventory || 10;

    // Handle input change
    input.addEventListener("change", async () => {
      let newQty = parseInt(input.value);
      if (isNaN(newQty) || newQty < 1) {
        newQty = 1;
        input.value = 1;
      }
      if (newQty > maxQuantity) {
        newQty = maxQuantity;
        input.value = maxQuantity;
      }

      try {
        // Remove cartId parameter since we use /carts/me endpoint now
        await cart.updateCartItem(null, itemId, newQty);

        // Update the item subtotal
        const cartItem = input.closest(".cart-item");
        const subtotalEl = cartItem.querySelector(".mt-3 .font-medium");
        if (subtotalEl) {
          subtotalEl.textContent = formatPrice(newQty * price);
        }

        // Update cart total
        updateCartTotal();
      } catch (error) {
        console.error("Failed to update quantity:", error);
        alert("Failed to update quantity");
      }
    });
  });

  // Handle decrease buttons
  document.querySelectorAll(".decrease-quantity-btn").forEach((button) => {
    const itemId = button.dataset.itemId;
    const input = document.querySelector(
      `.quantity-input[data-item-id="${itemId}"]`
    );
    if (!input) return;

    // Find the corresponding cart item in our data
    const itemData = cartItems.find(
      (item) => item.id === itemId || item.item_id === itemId
    );
    if (!itemData) return;

    const price = itemData.product?.price || itemData.unit_price || 0;

    button.addEventListener("click", async () => {
      let currentQty = parseInt(input.value);
      if (currentQty <= 1) return;

      const newQty = currentQty - 1;
      input.value = newQty;

      try {
        // Remove cartId parameter since we use /carts/me endpoint now
        await cart.updateCartItem(null, itemId, newQty);

        // Update the item subtotal
        const cartItem = button.closest(".cart-item");
        const subtotalEl = cartItem.querySelector(".mt-3 .font-medium");
        if (subtotalEl) {
          subtotalEl.textContent = formatPrice(newQty * price);
        }

        // Update cart total
        updateCartTotal();
      } catch (error) {
        console.error("Failed to decrease quantity:", error);
        input.value = currentQty; // Revert on error
        alert("Failed to update quantity");
      }
    });
  });

  // Handle increase buttons
  document.querySelectorAll(".increase-quantity-btn").forEach((button) => {
    const itemId = button.dataset.itemId;
    const input = document.querySelector(
      `.quantity-input[data-item-id="${itemId}"]`
    );
    if (!input) return;

    // Find the corresponding cart item in our data
    const itemData = cartItems.find(
      (item) => item.id === itemId || item.item_id === itemId
    );
    if (!itemData) return;

    const price = itemData.product?.price || itemData.unit_price || 0;
    const maxQuantity = itemData.product?.inventory || 10;

    button.addEventListener("click", async () => {
      let currentQty = parseInt(input.value);
      if (currentQty >= maxQuantity) return;

      const newQty = currentQty + 1;
      input.value = newQty;

      try {
        // Remove cartId parameter since we use /carts/me endpoint now
        await cart.updateCartItem(null, itemId, newQty);

        // Update the item subtotal
        const cartItem = button.closest(".cart-item");
        const subtotalEl = cartItem.querySelector(".mt-3 .font-medium");
        if (subtotalEl) {
          subtotalEl.textContent = formatPrice(newQty * price);
        }

        // Update cart total
        updateCartTotal();
      } catch (error) {
        console.error("Failed to increase quantity:", error);
        input.value = currentQty; // Revert on error
        alert("Failed to update quantity");
      }
    });
  });

  // Handle remove buttons
  document.querySelectorAll(".remove-item-btn").forEach((button) => {
    const itemId = button.dataset.itemId;

    button.addEventListener("click", async () => {
      if (
        confirm("Are you sure you want to remove this item from your cart?")
      ) {
        try {
          // Remove cartId parameter since we use /carts/me endpoint now
          await cart.removeFromCart(null, itemId);

          // Remove the item from the UI
          const cartItem = button.closest(".cart-item");
          cartItem.remove();

          // Update cart total
          updateCartTotal();

          // If no items left, show empty cart message
          if (document.querySelectorAll(".cart-item").length === 0) {
            const cartItemsContainer = document.getElementById("cart-items");
            if (cartItemsContainer) {
              cartItemsContainer.innerHTML = `
                <div class="p-8 text-center">
                  <div class="text-gray-500 mb-4"><i class="fas fa-shopping-cart text-4xl mb-4"></i></div>
                  <p class="text-lg text-gray-700 mb-4">Keranjang anda masih kosong</p>
                  <a href="products.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
                    <i class="fas fa-shopping-bag mr-2"></i> Lihat Produk
                  </a>
                </div>
              `;

              // Reset summary
              updateCartSummary(0, 0, 0);
            }
          }
        } catch (error) {
          console.error("Failed to remove item:", error);
          alert("Failed to remove item from cart");
        }
      }
    });
  });
}

// Debug function to help identify issues with checkout
function debugCart() {
  try {
    // Check for current user
    const user = auth.getCurrentUser();
    console.log("Current user:", user);

    // Try to get current cart
    fetch(`${API_BASE_URL}/carts/me`, {
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${localStorage.getItem("token")}`,
      },
    })
      .then((response) => response.json())
      .then((data) => {
        console.log("Current user cart:", data);

        // Check cart items if available
        if (data.data && data.data.id) {
          const cartId = data.data.id;

          // Try to inspect individual cart items
          if (data.data.items && data.data.items.length > 0) {
            console.log("Cart items:", data.data.items);

            // Check for potential soft-deleted items
            const itemIds = data.data.items.map((item) => item.id);
            console.log("Item IDs to check:", itemIds);
          }

          // Try to diagnose checkout issues
          console.log(
            "Attempting to diagnose checkout issues for cart:",
            cartId
          );

          // Test checkout API directly
          fetch(`${API_BASE_URL}/checkouts`, {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${localStorage.getItem("token")}`,
            },
            body: JSON.stringify({
              cart_id: cartId,
            }),
          })
            .then((checkoutResponse) => {
              console.log("Checkout API test status:", checkoutResponse.status);
              return checkoutResponse.json();
            })
            .then((checkoutData) => {
              console.log("Checkout API test response:", checkoutData);
            })
            .catch((error) => {
              console.error("Checkout API test error:", error);
            });
        }
      })
      .catch((error) => {
        console.error("Cart debug error:", error);
      });
  } catch (error) {
    console.error("Debug function error:", error);
  }
}

// Create a global debug function
window.debugCartIssues = debugCart;
