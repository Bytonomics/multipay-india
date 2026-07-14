package cashfree

import (
	"math"

	cf "github.com/cashfree/cashfree-pg/v6"

	"github.com/Bytonomics/multipay-india/multipay-go/domain"
	"github.com/Bytonomics/multipay-india/multipay-go/utils/currencyutils"
)

// This file maps the richer canonical Create-Order optional structures (cart,
// terminal, splits, products, payment-method filters) onto their Cashfree SDK
// structs. Money fields are converted from minor units to Cashfree's major
// (float) wire format via currencyutils, threading the order currency.

// mapCartDetails maps canonical CartDetails → cf.CartDetails. currencyCode is the
// order currency, used to convert minor money fields to major units.
func mapCartDetails(c *domain.CartDetails, currencyCode string) *cf.CartDetails {
	out := &cf.CartDetails{}
	if c.CustomerNote != "" {
		v := c.CustomerNote
		out.CustomerNote = &v
	}
	if c.ShippingCharge != 0 {
		v := currencyutils.AmountMinorToMajor(int64(c.ShippingCharge), currencyCode)
		out.ShippingCharge = &v
	}
	if c.CartName != "" {
		v := c.CartName
		out.CartName = &v
	}
	if c.ShippingAddress != nil {
		out.CustomerShippingAddress = mapCartAddress(c.ShippingAddress)
	}
	if c.BillingAddress != nil {
		out.CustomerBillingAddress = mapCartAddress(c.BillingAddress)
	}
	if len(c.CartItems) > 0 {
		items := make([]cf.CartItem, 0, len(c.CartItems))
		for i := range c.CartItems {
			items = append(items, mapCartItem(&c.CartItems[i], currencyCode))
		}
		out.CartItems = items
	}
	return out
}

// mapCartAddress maps canonical CartAddress → cf.CartAddress.
func mapCartAddress(a *domain.CartAddress) *cf.CartAddress {
	out := &cf.CartAddress{}
	if a.FullName != "" {
		v := a.FullName
		out.FullName = &v
	}
	if a.Country != "" {
		v := a.Country
		out.Country = &v
	}
	if a.City != "" {
		v := a.City
		out.City = &v
	}
	if a.State != "" {
		v := a.State
		out.State = &v
	}
	if a.Pincode != "" {
		v := a.Pincode
		out.Pincode = &v
	}
	if a.Address1 != "" {
		v := a.Address1
		out.Address1 = &v
	}
	if a.Address2 != "" {
		v := a.Address2
		out.Address2 = &v
	}
	return out
}

// mapCartItem maps canonical CartItem → cf.CartItem. Prices are converted from
// minor to major using the item currency if set, else the order currency.
func mapCartItem(i *domain.CartItem, orderCurrency string) cf.CartItem {
	itemCurrency := orderCurrency
	if i.ItemCurrency != "" {
		itemCurrency = string(i.ItemCurrency)
	}
	out := cf.CartItem{}
	if i.ItemID != "" {
		v := i.ItemID
		out.ItemId = &v
	}
	if i.ItemName != "" {
		v := i.ItemName
		out.ItemName = &v
	}
	if i.ItemDescription != "" {
		v := i.ItemDescription
		out.ItemDescription = &v
	}
	if len(i.ItemTags) > 0 {
		out.ItemTags = i.ItemTags
	}
	if i.ItemDetailsURL != "" {
		v := i.ItemDetailsURL
		out.ItemDetailsUrl = &v
	}
	if i.ItemImageURL != "" {
		v := i.ItemImageURL
		out.ItemImageUrl = &v
	}
	if i.OriginalUnitPrice != 0 {
		v := currencyutils.AmountMinorToMajor(int64(i.OriginalUnitPrice), itemCurrency)
		out.ItemOriginalUnitPrice = &v
	}
	if i.DiscountedUnitPrice != 0 {
		v := currencyutils.AmountMinorToMajor(int64(i.DiscountedUnitPrice), itemCurrency)
		out.ItemDiscountedUnitPrice = &v
	}
	if i.ItemCurrency != "" {
		v := string(i.ItemCurrency)
		out.ItemCurrency = &v
	}
	if i.ItemQuantity != 0 {
		v := float32(i.ItemQuantity)
		out.ItemQuantity = &v
	}
	if i.ItemVariantID != "" {
		v := i.ItemVariantID
		out.ItemVariantId = &v
	}
	return out
}

// mapTerminalDetails maps canonical TerminalDetails → cf.TerminalDetails. Only
// TerminalType is a mandatory field on the vendor struct.
func mapTerminalDetails(t *domain.TerminalDetails) *cf.TerminalDetails {
	out := &cf.TerminalDetails{
		TerminalType: t.TerminalType,
	}
	if t.TerminalID != "" {
		v := t.TerminalID
		out.TerminalId = &v
	}
	if t.TerminalPhoneNo != "" {
		v := t.TerminalPhoneNo
		out.TerminalPhoneNo = &v
	}
	if t.TerminalName != "" {
		v := t.TerminalName
		out.TerminalName = &v
	}
	if t.CfTerminalID != 0 {
		v := t.CfTerminalID
		out.CfTerminalId = &v
	}
	if t.TerminalAddress != "" {
		v := t.TerminalAddress
		out.TerminalAddress = &v
	}
	if t.TerminalNote != "" {
		v := t.TerminalNote
		out.TerminalNote = &v
	}
	return out
}

// mapVendorSplits maps canonical []VendorSplit → []cf.VendorSplit. Amount is
// converted from minor to major units; Tags are wrapped in Cashfree's nested map.
func mapVendorSplits(splits []domain.VendorSplit, currencyCode string) []cf.VendorSplit {
	out := make([]cf.VendorSplit, 0, len(splits))
	for i := range splits {
		s := &splits[i]
		cfSplit := cf.VendorSplit{}
		if s.VendorID != "" {
			v := s.VendorID
			cfSplit.VendorId = &v
		}
		if s.Amount != 0 {
			v := float32(currencyutils.AmountMinorToMajor(int64(s.Amount), currencyCode))
			cfSplit.Amount = &v
		}
		if s.Percentage != 0 {
			v := float32(s.Percentage)
			cfSplit.Percentage = &v
		}
		if len(s.Tags) > 0 {
			cfSplit.Tags = wrapSplitTags(s.Tags)
		}
		out = append(out, cfSplit)
	}
	return out
}

// wrapSplitTags converts a flat string map into Cashfree's split tags shape
// (map[string]map[string]interface{}). Each value is stored under a "value" key.
func wrapSplitTags(tags map[string]string) map[string]map[string]interface{} {
	out := make(map[string]map[string]interface{}, len(tags))
	for k, v := range tags {
		out[k] = map[string]interface{}{"value": v}
	}
	return out
}

// mapOrderProducts maps canonical OrderProducts → cf.Products. A non-nil pointer
// on the canonical side means "enable this product".
func mapOrderProducts(p *domain.OrderProducts) *cf.Products {
	out := &cf.Products{}
	if p.OneClickCheckout != nil {
		out.OneClickCheckout = &cf.ProductDetails{}
	}
	if p.VerifyPay != nil {
		out.VerifyPay = &cf.ProductDetails{}
	}
	return out
}

// mapPaymentMethodsFilters maps canonical OrderPaymentMethodsFilters →
// cf.OrderMetaPaymentMethodsFilters (methods + card bin/scheme/bank/suffix filters).
func mapPaymentMethodsFilters(f *domain.OrderPaymentMethodsFilters) *cf.OrderMetaPaymentMethodsFilters {
	out := &cf.OrderMetaPaymentMethodsFilters{}
	if f.Methods != nil {
		methods := &cf.OrderMetaPaymentMethodsFiltersMethods{
			Values: f.Methods.Values,
		}
		if f.Methods.Action != "" {
			action := f.Methods.Action
			methods.Action = &action
		}
		out.Methods = methods
	}
	if f.Filters != nil {
		filters := &cf.OrderPaymentMethodFilters{}
		if len(f.Filters.CardBins) > 0 {
			filters.CardBins = int64SliceToInt32(f.Filters.CardBins)
		}
		if len(f.Filters.CardSchemes) > 0 {
			filters.CardSchemes = f.Filters.CardSchemes
		}
		if len(f.Filters.CardSuffix) > 0 {
			filters.CardSuffix = int64SliceToInt32(f.Filters.CardSuffix)
		}
		if len(f.Filters.CardIssuingBank) > 0 {
			filters.CardIssuingBank = f.Filters.CardIssuingBank
		}
		out.Filters = filters
	}
	return out
}

// int64SliceToInt32 narrows a canonical []int64 to the Cashfree wire []int32.
// Values outside the int32 range are clamped to the int32 bounds so a stray
// out-of-range bin/suffix can never overflow into a bogus negative wire value.
func int64SliceToInt32(in []int64) []int32 {
	out := make([]int32, 0, len(in))
	for _, v := range in {
		switch {
		case v > math.MaxInt32:
			out = append(out, math.MaxInt32)
		case v < math.MinInt32:
			out = append(out, math.MinInt32)
		default:
			out = append(out, int32(v))
		}
	}
	return out
}
