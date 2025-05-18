/**
 * Checkout page functionality
 */

// Initialize checkout page
function initCheckoutPage() {
  const checkoutContainer = document.querySelector(".checkout-container");
  if (!checkoutContainer) return;

  // Show loading state
  checkoutContainer.innerHTML =
    '<div class="loading-spinner">Loading checkout...</div>';

  try {
    // Load the user's cart first
    loadCheckoutData(checkoutContainer);
  } catch (error) {
    console.error("Error processing checkout:", error);
    checkoutContainer.innerHTML = `<p class="text-center">Error processing checkout: ${error.message}</p>`;
  }
}

// Load checkout data
async function loadCheckoutData(container) {
  try {
    // Get user's cart
    const cartResponse = await cart.getCart();
    if (!cartResponse || !cartResponse.data) {
      container.innerHTML = `
        <div class="p-8 text-center">
          <div class="text-yellow-500 mb-4"><i class="fas fa-exclamation-triangle text-4xl mb-4"></i></div>
          <p class="text-lg text-gray-700 mb-4">Your cart is empty. Add some products before checkout.</p>
          <a href="products.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
            <i class="fas fa-shopping-bag mr-2"></i> View Products
          </a>
        </div>
      `;
      return;
    }

    const userCart = cartResponse.data;

    // Extract cart items
    let cartItems = [];
    if (userCart.items && Array.isArray(userCart.items)) {
      cartItems = userCart.items;
    } else if (userCart.cart_items && Array.isArray(userCart.cart_items)) {
      cartItems = userCart.cart_items;
    }

    if (cartItems.length === 0) {
      container.innerHTML = `
        <div class="p-8 text-center">
          <div class="text-yellow-500 mb-4"><i class="fas fa-exclamation-triangle text-4xl mb-4"></i></div>
          <p class="text-lg text-gray-700 mb-4">Your cart is empty. Add some products before checkout.</p>
          <a href="products.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
            <i class="fas fa-shopping-bag mr-2"></i> View Products
          </a>
        </div>
      `;
      return;
    }

    // Prepare items for checkout display and calculation
    const checkoutItems = cartItems.map((item) => {
      const product = item.product || {};
      return {
        product_id: product.id || item.product_id,
        product_name: product.name || item.product_name || "Unnamed Product",
        product_sku: product.sku || item.product_sku || "",
        unit_price: product.price || item.unit_price || 0,
        quantity: item.quantity || 1,
        subtotal:
          (item.quantity || 1) * (product.price || item.unit_price || 0),
      };
    });

    // Apply promotions according to business rules
    const checkout = applyPromotions(checkoutItems);

    // Render the checkout page
    renderCheckout(checkout, container);
  } catch (error) {
    console.error("Error loading checkout data:", error);
    container.innerHTML = `
      <div class="p-8 text-center">
        <div class="text-red-500 mb-4"><i class="fas fa-exclamation-circle text-4xl mb-4"></i></div>
        <p class="text-lg text-gray-700 mb-4">Error loading checkout: ${error.message}</p>
        <a href="cart.html" class="inline-flex items-center bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md">
          <i class="fas fa-shopping-cart mr-2"></i> Return to Cart
        </a>
      </div>
    `;
  }
}

// Function to generate a UUID (for the demo)
function uuid() {
  return "xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx".replace(/[xy]/g, function (c) {
    var r = (Math.random() * 16) | 0,
      v = c === "x" ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

// Apply promotions to cart items according to the test requirements
function applyPromotions(cartItems) {
  const checkout = {
    id: uuid(),
    items: [...cartItems],
    promotions: [],
    subtotal: 0,
    total_discount: 0,
    total: 0,
  };

  // Calculate initial subtotal
  checkout.subtotal = cartItems.reduce(
    (total, item) => total + item.subtotal,
    0
  );
  checkout.total = checkout.subtotal;

  // 1. Each sale of a MacBook Pro comes with a free Raspberry Pi B
  const macbookPromotion = applyMacbookPromotion(cartItems);
  if (macbookPromotion.discount > 0) {
    checkout.promotions.push(macbookPromotion);
    checkout.total_discount += macbookPromotion.discount;
  }

  // 2. Buy 3 Google Homes for the price of 2
  const googleHomePromotion = applyGoogleHomePromotion(cartItems);
  if (googleHomePromotion.discount > 0) {
    checkout.promotions.push(googleHomePromotion);
    checkout.total_discount += googleHomePromotion.discount;
  }

  // 3. Buying more than 3 Alexa Speakers will get a 10% discount on all Alexa speakers
  const alexaPromotion = applyAlexaPromotion(cartItems);
  if (alexaPromotion.discount > 0) {
    checkout.promotions.push(alexaPromotion);
    checkout.total_discount += alexaPromotion.discount;
  }

  // Calculate final total
  checkout.total = checkout.subtotal - checkout.total_discount;

  return checkout;
}

// Apply MacBook Pro promotion - each MacBook Pro comes with a free Raspberry Pi B
function applyMacbookPromotion(cartItems) {
  const macbooks = cartItems.find((item) => item.product_sku === "43N23P");
  const raspberryPis = cartItems.find((item) => item.product_sku === "234234");

  const promotion = {
    id: uuid(),
    description: "Each sale of a MacBook Pro comes with a free Raspberry Pi B",
    discount: 0,
  };

  if (!macbooks || !raspberryPis) {
    return promotion;
  }

  // The number of free Raspberry Pis equals the number of MacBooks purchased
  // but is limited by the number of Raspberry Pis in the cart
  const freeItems = Math.min(macbooks.quantity, raspberryPis.quantity);

  if (freeItems > 0) {
    promotion.discount = freeItems * raspberryPis.unit_price;
  }

  return promotion;
}

// Apply Google Home promotion - Buy 3 Google Homes for the price of 2
function applyGoogleHomePromotion(cartItems) {
  const googleHomes = cartItems.find((item) => item.product_sku === "120P90");

  const promotion = {
    id: uuid(),
    description: "Buy 3 Google Homes for the price of 2",
    discount: 0,
  };

  if (!googleHomes || googleHomes.quantity < 3) {
    return promotion;
  }

  // For every 3 Google Homes, 1 is free
  const sets = Math.floor(googleHomes.quantity / 3);
  promotion.discount = sets * googleHomes.unit_price;

  return promotion;
}

// Apply Alexa Speaker promotion - Buying more than 3 gets a 10% discount on all
function applyAlexaPromotion(cartItems) {
  const alexaSpeakers = cartItems.find((item) => item.product_sku === "A304SD");

  const promotion = {
    id: uuid(),
    description: "10% discount on Alexa Speakers when buying more than 3",
    discount: 0,
  };

  if (!alexaSpeakers || alexaSpeakers.quantity <= 3) {
    return promotion;
  }

  // 10% discount on all Alexa speakers
  promotion.discount = alexaSpeakers.subtotal * 0.1;

  return promotion;
}

// Render the checkout with applied promotions
function renderCheckout(checkout, container) {
  const itemsHtml = checkout.items
    .map((item) => {
      return `
      <div class="checkout-item p-4 border-b border-gray-200">
        <div class="flex justify-between">
          <div>
            <h3 class="font-medium">${item.product_name}</h3>
            ${
              item.product_sku
                ? `<p class="text-sm text-gray-500">SKU: ${item.product_sku}</p>`
                : ""
            }
            <p class="text-sm">Price: ${formatPrice(item.unit_price)}</p>
            <p class="text-sm">Quantity: ${item.quantity}</p>
          </div>
          <div class="text-right">
            <p class="font-medium">${formatPrice(item.subtotal)}</p>
          </div>
        </div>
      </div>
      `;
    })
    .join("");

  const promotionsHtml = checkout.promotions.length
    ? checkout.promotions
        .map(
          (promotion) => `
          <div class="flex justify-between p-2 bg-green-50 border-b border-green-100">
            <p class="text-sm text-green-800">${promotion.description}</p>
            <p class="text-sm font-medium text-green-800">-${formatPrice(
              promotion.discount
            )}</p>
          </div>
        `
        )
        .join("")
    : `<p class="text-sm text-gray-500 p-2">No promotions applied</p>`;

  container.innerHTML = `
    <div class="bg-white shadow-sm rounded-md overflow-hidden">
      <div class="bg-gray-50 px-4 py-3 border-b border-gray-200">
        <h2 class="text-lg font-medium text-gray-800">Order Summary</h2>
      </div>
      
      <div class="divide-y divide-gray-200">
        ${itemsHtml}
      </div>
      
      <div class="bg-gray-50 px-4 py-3 border-b border-gray-200">
        <h3 class="text-md font-medium text-gray-800">Applied Promotions</h3>
      </div>
      
      <div>
        ${promotionsHtml}
      </div>
      
      <div class="p-4 space-y-2">
        <div class="flex justify-between">
          <p class="text-gray-600">Subtotal:</p>
          <p class="font-medium">${formatPrice(checkout.subtotal)}</p>
        </div>
        
        <div class="flex justify-between text-green-700">
          <p>Discount:</p>
          <p>-${formatPrice(checkout.total_discount)}</p>
        </div>
        
        <div class="flex justify-between text-lg font-medium border-t border-gray-200 pt-2 mt-2">
          <p>Total:</p>
          <p>${formatPrice(checkout.total)}</p>
        </div>
      </div>
      
      <div class="p-4 bg-gray-50 border-t border-gray-200">
        <div class="flex flex-col space-y-3 sm:flex-row sm:space-y-0 sm:space-x-3">
          <a href="cart.html" class="inline-block text-center px-4 py-2 border border-gray-300 rounded-md font-medium text-gray-700 bg-white hover:bg-gray-50">
            Back to Cart
          </a>
          <button id="place-order-btn" class="inline-block text-center px-4 py-2 border border-transparent rounded-md font-medium text-white bg-green-600 hover:bg-green-700">
            Place Order
          </button>
        </div>
      </div>
    </div>
  `;

  // Add event listener for place order button
  const placeOrderBtn = container.querySelector("#place-order-btn");
  if (placeOrderBtn) {
    placeOrderBtn.addEventListener("click", async () => {
      try {
        placeOrderBtn.textContent = "Processing...";
        placeOrderBtn.disabled = true;

        // Simulate order processing
        await new Promise((resolve) => setTimeout(resolve, 1500));

        alert("Your order has been placed successfully!");
        window.location.href = "products.html";
      } catch (error) {
        console.error("Error placing order:", error);
        alert("Failed to place order: " + error.message);
        placeOrderBtn.textContent = "Place Order";
        placeOrderBtn.disabled = false;
      }
    });
  }
}
