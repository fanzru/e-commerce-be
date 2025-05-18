/**
 * Authentication functionality for login and register pages
 */

// Initialize login page
function initLoginPage() {
  console.log("Initializing login page");
  const loginForm = document.getElementById("login-form");
  if (!loginForm) {
    console.error("Login form not found in the DOM");
    return;
  }

  // Check if Tailwind CSS might be loaded
  try {
    if (typeof window.tailwind === "undefined") {
      console.warn("Tailwind CSS might not be loaded correctly");
    }
  } catch (e) {
    console.warn("Could not check for Tailwind CSS:", e);
  }

  // Add a status element for feedback if it doesn't exist
  let statusDiv = document.getElementById("login-status");
  if (!statusDiv) {
    statusDiv = document.createElement("div");
    statusDiv.id = "login-status";
    statusDiv.className = "mt-4 p-3 rounded-md hidden";
    loginForm.appendChild(statusDiv);
  }

  loginForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    console.log("Login form submitted");

    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;

    // Basic validation
    if (!email) {
      showStatus(statusDiv, "Please enter your email address", "error");
      return;
    }

    if (!validateEmail(email)) {
      showStatus(statusDiv, "Please enter a valid email address", "error");
      return;
    }

    if (!password) {
      showStatus(statusDiv, "Please enter your password", "error");
      return;
    }

    // Show loading status
    showStatus(statusDiv, "Logging in...", "loading");

    try {
      const result = await auth.login(email, password);
      console.log("Login successful, redirecting to products page");

      // Show success message
      showStatus(statusDiv, "Login successful! Redirecting...", "success");

      // Redirect after a short delay
      setTimeout(() => {
        window.location.href = "products.html";
      }, 1000);
    } catch (error) {
      console.error("Login error:", error);
      showStatus(
        statusDiv,
        error.message ||
          "Login failed. Please check your credentials and try again.",
        "error"
      );
    }
  });
}

// Initialize register page
function initRegisterPage() {
  console.log("Initializing register page");
  const registerForm = document.getElementById("register-form");
  if (!registerForm) {
    console.error("Register form not found in the DOM");
    return;
  }

  // Check if Tailwind CSS might be loaded
  try {
    if (typeof window.tailwind === "undefined") {
      console.warn("Tailwind CSS might not be loaded correctly");
    }
  } catch (e) {
    console.warn("Could not check for Tailwind CSS:", e);
  }

  // Add a status element for feedback if it doesn't exist
  let statusDiv = document.getElementById("register-status");
  if (!statusDiv) {
    statusDiv = document.createElement("div");
    statusDiv.id = "register-status";
    statusDiv.className = "mt-4 p-3 rounded-md hidden";
    registerForm.appendChild(statusDiv);
  }

  registerForm.addEventListener("submit", async (e) => {
    e.preventDefault();
    console.log("Register form submitted");

    const name = document.getElementById("name").value.trim();
    const email = document.getElementById("email").value.trim();
    const password = document.getElementById("password").value;

    // Basic validation
    if (!name) {
      showStatus(statusDiv, "Please enter your name", "error");
      return;
    }

    if (!email) {
      showStatus(statusDiv, "Please enter your email address", "error");
      return;
    }

    if (!validateEmail(email)) {
      showStatus(statusDiv, "Please enter a valid email address", "error");
      return;
    }

    if (!password) {
      showStatus(statusDiv, "Please enter a password", "error");
      return;
    }

    if (password.length < 6) {
      showStatus(
        statusDiv,
        "Password must be at least 6 characters long",
        "error"
      );
      return;
    }

    // Show loading status
    showStatus(statusDiv, "Creating your account...", "loading");

    try {
      await auth.register({ name, email, password });
      console.log("Registration successful");

      // Show success message
      showStatus(
        statusDiv,
        "Registration successful! Redirecting to login...",
        "success"
      );

      // Redirect after a short delay
      setTimeout(() => {
        window.location.href = "login.html";
      }, 1500);
    } catch (error) {
      console.error("Registration error:", error);
      showStatus(
        statusDiv,
        error.message || "Registration failed. Please try again.",
        "error"
      );
    }
  });
}

// Helper function to show status messages
function showStatus(element, message, type) {
  if (!element) return;

  // Make element visible
  element.style.display = "block";
  element.textContent = message;

  // Reset classes
  element.className = "mt-4 p-3 rounded-md";

  // Apply styling based on message type
  switch (type) {
    case "error":
      element.style.backgroundColor = "#f8d7da";
      element.style.color = "#721c24";
      break;
    case "success":
      element.style.backgroundColor = "#d4edda";
      element.style.color = "#155724";
      break;
    case "loading":
      element.style.backgroundColor = "#f8f9fa";
      element.style.color = "#333";
      break;
    default:
      element.style.backgroundColor = "#f8f9fa";
      element.style.color = "#333";
  }
}
