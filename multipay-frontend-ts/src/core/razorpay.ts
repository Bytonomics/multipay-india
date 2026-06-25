import type { RazorpayCheckoutPayload, RazorpayFormFields } from "./types";

export function checkoutRazorpay(payload: RazorpayCheckoutPayload): void {
  const form = document.createElement("form");
  form.method = "POST";
  form.action = "https://api.razorpay.com/v1/checkout/embedded";

  const fields: RazorpayFormFields = {
    key_id: payload.public_key,
    order_id: payload.order_id,
    amount: String(payload.amount_minor),
    currency: payload.currency,
    callback_url: payload.callback_url || "",
  };

  for (const [key, value] of Object.entries(fields)) {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = key;
    input.value = value;
    form.appendChild(input);
  }

  document.body.appendChild(form);
  form.submit();
}
