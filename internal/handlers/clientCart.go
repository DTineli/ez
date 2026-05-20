package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/DTineli/ez/internal/middleware"
	"github.com/DTineli/ez/internal/store"
	"gorm.io/gorm"
)

func (c *ClientHandler) PostAddToCart(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		ShowToast(w, "Dados invalidos", "error")
		return
	}

	productID, err := formUint(r, "product_id")
	if err != nil {
		ShowToast(w, "Produto invalido", "error")
		return
	}

	variantID, err := formUint(r, "variant_id")
	if err != nil {
		ShowToast(w, "Variacao invalida", "error")
		return
	}

	qty, err := formPosInt(r, "qty")
	if err != nil {
		ShowToast(w, "Quantidade invalida", "error")
		return
	}

	sess := middleware.GetSessionFromContext(r)
	variant, err := c.productStore.GetVariantForCart(
		uint(variantID),
		uint(productID),
		sess.TenantID,
	)

	if err != nil {
		ShowToast(w, "Produto ou variacao invalida", "error")
		return
	}

	price := variant.CostPrice
	cart, err := c.resolveOpenCart(r, w, sess)
	if err != nil {
		ShowToast(w, "Erro ao preparar carrinho", "error")
		return
	}

	if err := c.cartStore.AddOrIncrementItem(cart.ID, variant.ProductID, variant.ID, qty, price); err != nil {
		ShowToast(w, "Erro ao adicionar item", "error")
		return
	}

	totalItems, err := c.cartStore.CountItems(cart.ID)
	if err != nil {
		ShowToast(w, "Produto adicionado ao carrinho", "success")
		return
	}

	w.Header().Set("HX-Trigger", fmt.Sprintf(`{
	"showToast": {
		"type": "success",
		"message": "Produto adicionado ao carrinho"
	},
	"cartCountUpdated": {
		"count": %d
	}
}`, totalItems))
	w.WriteHeader(http.StatusOK)
}

func (c *ClientHandler) resolveOpenCart(
	r *http.Request,
	w http.ResponseWriter,
	sess *store.Session,
) (*store.Cart, error) {
	// tentei achar o carrinho
	if sess.CartID != 0 {
		cart, err := c.cartStore.FindOpenByID(
			sess.CartID,
			sess.TenantID,
			sess.ContactInfo.ID,
		)
		if err == nil {
			return cart, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// dai eu procuro por contato
	cart, err := c.cartStore.FindOpenByContact(
		sess.TenantID,
		sess.ContactInfo.ID,
	)
	// crio carrinho na sessao
	if err == nil {
		if setErr := c.sessionStore.SetCartID(r, w, cart.ID); setErr != nil {
			return nil, setErr
		}
		return cart, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	//se nao eu crio um novo
	newCart := &store.Cart{
		TenantID:  sess.TenantID,
		ContactID: sess.ContactInfo.ID,
		Status:    store.CartStatusOpen,
	}
	if err := c.cartStore.Create(newCart); err != nil {
		return nil, err
	}

	if err := c.sessionStore.SetCartID(r, w, newCart.ID); err != nil {
		return nil, err
	}

	return newCart, nil
}

func (c *ClientHandler) DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	productID, ok := parseURLParamUint(w, r, "productID", "Item invalido")
	if !ok {
		return
	}

	variantID, ok := parseURLParamUint(w, r, "variantID", "Item invalido")
	if !ok {
		return
	}

	sess := middleware.GetSessionFromContext(r)
	cart, err := c.resolveOpenCart(r, w, sess)
	if err != nil {
		ShowToast(w, "Erro ao acessar carrinho", "error")
		return
	}

	if err := c.cartStore.RemoveItem(cart.ID, uint(productID), uint(variantID)); err != nil {
		ShowToast(w, "Erro ao remover item", "error")
		return
	}

	w.Header().Set(HXRedirect, "/client/confirmacao")
	w.WriteHeader(http.StatusOK)
}

func (c *ClientHandler) PatchCartItemQty(
	w http.ResponseWriter,
	r *http.Request,
) {
	productID, ok := parseURLParamUint(w, r, "productID", "Item invalido")
	if !ok {
		return
	}

	variantID, ok := parseURLParamUint(w, r, "variantID", "Item invalido")
	if !ok {
		return
	}

	if err := r.ParseForm(); err != nil {
		ShowToast(w, "Dados invalidos", "error")
		return
	}

	qty, err := formPosInt(r, "qty")
	if err != nil {
		ShowToast(w, "Quantidade invalida", "error")
		return
	}

	sess := middleware.GetSessionFromContext(r)
	cart, err := c.resolveOpenCart(r, w, sess)
	if err != nil {
		ShowToast(w, "Erro ao acessar carrinho", "error")
		return
	}

	if err := c.cartStore.UpdateItemQty(
		cart.ID,
		uint(productID),
		uint(variantID),
		qty); err != nil {
		ShowToast(w, "Erro ao atualizar quantidade", "error")
		return
	}

	w.Header().Set(HXRedirect, "/client/confirmacao")
	w.WriteHeader(http.StatusOK)
}

func (c *ClientHandler) PostConfirmOrder(
	w http.ResponseWriter,
	r *http.Request,
) {
	if err := r.ParseForm(); err != nil {
		ShowToast(w, "Dados invalidos", "error")
		return
	}

	var priceTableID uint
	if v, err := strconv.ParseUint(r.FormValue("price_table"), 10, 64); err == nil {
		priceTableID = uint(v)
	}

	sess := middleware.GetSessionFromContext(r)

	var cart *store.Cart
	var err error

	if sess.CartID != 0 {
		cart, err = c.cartStore.FindOpenByID(
			sess.CartID,
			sess.TenantID,
			sess.ContactInfo.ID,
		)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			ShowToast(w, "Erro ao preparar carrinho", "error")
			return
		}
	}

	if cart == nil {
		cart, err = c.cartStore.FindOpenByContact(
			sess.TenantID,
			sess.ContactInfo.ID,
		)
		if err != nil {
			ShowToast(w, "Carrinho vazio", "error")
			return
		}
		_ = c.sessionStore.SetCartID(r, w, cart.ID)
	}

	_, err = c.orderStore.ConfirmFromCart(
		cart.ID,
		sess.TenantID,
		sess.ContactInfo.ID,
		priceTableID,
	)

	if err != nil {
		ShowToast(w, "Erro ao confirmar pedido", "error")
		return
	}

	_ = c.sessionStore.SetCartID(r, w, 0)
	w.Header().Set(HXRedirect, "/client/items")
	w.WriteHeader(http.StatusOK)
}
