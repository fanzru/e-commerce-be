/**
 * Main JavaScript entry point
 * This file loads all needed scripts
 */

// Wait for DOM to be fully loaded before executing our code
document.addEventListener("DOMContentLoaded", function () {
  console.log("E-Commerce frontend initialized");
});

// Add product-placeholder SVG if it doesn't exist
if (!document.getElementById("product-placeholder-svg")) {
  const placeholderSvg = document.createElement("div");
  placeholderSvg.id = "product-placeholder-svg";
  placeholderSvg.innerHTML = `
    <svg xmlns="http://www.w3.org/2000/svg" width="0" height="0" style="display:none;">
      <symbol id="product-placeholder" viewBox="0 0 240 240">
        <rect width="240" height="240" fill="#f0f0f0"/>
        <path d="M120,80 L160,160 L80,160 Z" fill="#d0d0d0"/>
        <circle cx="160" cy="80" r="20" fill="#d0d0d0"/>
      </symbol>
    </svg>
  `;
  placeholderSvg.style.display = "none";
  document.body.appendChild(placeholderSvg);
}
