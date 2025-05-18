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
        <div class="flex justify-between items-center">
          <div class="flex items-center">
            <div class="bg-gray-100 w-16 h-16 flex-shrink-0 rounded-md flex items-center justify-center mr-4">
              <i class="fas fa-box text-gray-400"></i>
            </div>
            <div>
              <h3 class="font-medium text-gray-800">${item.product_name}</h3>
              ${
                item.product_sku
                  ? `<p class="text-xs text-gray-500">SKU: ${item.product_sku}</p>`
                  : ""
              }
              <div class="flex space-x-4 mt-1">
                <p class="text-sm text-gray-600">Price: ${formatPrice(
                  item.unit_price
                )}</p>
                <p class="text-sm text-gray-600">Qty: ${item.quantity}</p>
              </div>
            </div>
          </div>
          <div class="text-right">
            <p class="font-medium text-gray-800">${formatPrice(
              item.subtotal
            )}</p>
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
          <div class="flex justify-between items-center p-3 bg-green-50 border-b border-green-100">
            <div class="flex items-center">
              <span class="text-green-600 mr-2"><i class="fas fa-tag"></i></span>
              <p class="text-sm text-green-800">${promotion.description}</p>
            </div>
            <p class="text-sm font-medium text-green-800">-${formatPrice(
              promotion.discount
            )}</p>
          </div>
        `
        )
        .join("")
    : `<div class="p-4 text-center text-gray-500 bg-gray-50"><p class="text-sm">No promotions available for this order</p></div>`;

  container.innerHTML = `
    <div class="bg-white shadow-sm rounded-lg overflow-hidden">
      <div class="bg-gray-50 px-4 py-3 border-b border-gray-200">
        <h2 class="text-lg font-medium text-gray-800">Order Items</h2>
      </div>
      
      <div class="divide-y divide-gray-200">
        ${itemsHtml}
      </div>
      
      <div class="bg-gray-50 px-4 py-3 border-b border-gray-200 mt-4">
        <h3 class="text-md font-medium text-gray-800">Applied Promotions</h3>
      </div>
      
      <div>
        ${promotionsHtml}
      </div>
      
      <div class="p-6 space-y-3 bg-gray-50">
        <div class="flex justify-between">
          <p class="text-gray-600">Subtotal:</p>
          <p class="font-medium">${formatPrice(checkout.subtotal)}</p>
        </div>
        
        <div class="flex justify-between text-green-700 ${
          checkout.total_discount > 0 ? "" : "hidden"
        }">
          <p>Discount:</p>
          <p>-${formatPrice(checkout.total_discount)}</p>
        </div>
        
        <div class="flex justify-between text-lg font-medium border-t border-gray-200 pt-3 mt-3">
          <p>Total:</p>
          <p class="text-green-700">${formatPrice(checkout.total)}</p>
        </div>
      </div>
      
      <div class="p-6 bg-white border-t border-gray-200">
        <div class="mb-6">
          <h3 class="text-md font-medium text-gray-800 mb-3">Payment Method</h3>
          <div class="flex space-x-4">
            <label class="border rounded-md p-3 flex items-center cursor-pointer bg-blue-50 border-blue-300">
              <input type="radio" name="payment_method" value="credit_card" checked class="mr-2">
              <span><i class="fas fa-credit-card mr-2"></i> Credit Card</span>
            </label>
            <label class="border rounded-md p-3 flex items-center cursor-pointer">
              <input type="radio" name="payment_method" value="paypal" class="mr-2">
              <span><i class="fab fa-paypal mr-2"></i> PayPal</span>
            </label>
          </div>
        </div>

        <div class="flex flex-col space-y-3 sm:flex-row sm:space-y-0 sm:space-x-4">
          <a href="cart.html" class="inline-flex justify-center items-center px-4 py-2 border border-gray-300 rounded-md font-medium text-gray-700 bg-white hover:bg-gray-50 transition-colors duration-150">
            <i class="fas fa-arrow-left mr-2"></i> Back to Cart
          </a>
          <button id="place-order-btn" class="inline-flex justify-center items-center px-4 py-2 border border-transparent rounded-md font-medium text-white bg-green-600 hover:bg-green-700 transition-colors duration-150">
            <i class="fas fa-check-circle mr-2"></i> Place Order
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
        placeOrderBtn.innerHTML =
          '<i class="fas fa-spinner fa-spin mr-2"></i> Processing...';
        placeOrderBtn.disabled = true;

        // Simulate order processing
        await new Promise((resolve) => setTimeout(resolve, 1500));

        // Show success message
        container.innerHTML = `
          <div class="bg-white shadow-sm rounded-lg overflow-hidden p-8 text-center">
            <div class="text-green-500 mb-4">
              <i class="fas fa-check-circle text-5xl"></i>
            </div>
            <h2 class="text-2xl font-bold text-gray-800 mb-2">Order Placed Successfully!</h2>
            <p class="text-gray-600 mb-6">Thank you for your purchase. Your order is being processed.</p>
            <a href="products.html" class="inline-flex items-center justify-center px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 transition-colors">
              <i class="fas fa-shopping-bag mr-2"></i> Continue Shopping
            </a>
          </div>
        `;
      } catch (error) {
        console.error("Error placing order:", error);
        alert("Failed to place order: " + error.message);
        placeOrderBtn.innerHTML =
          '<i class="fas fa-check-circle mr-2"></i> Place Order';
        placeOrderBtn.disabled = false;
      }
    });
  }
}
