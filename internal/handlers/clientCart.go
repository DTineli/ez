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

	productID, err := strconv.ParseUint(r.FormValue("product_id"), 10, 64)
	if err != nil || productID == 0 {
		ShowToast(w, "Produto invalido", "error")
		return
	}

	qty, err := strconv.Atoi(r.FormValue("qty"))
	if err != nil || qty <= 0 {
		ShowToast(w, "Quantidade invalida", "error")
		return
	}

	sess := middleware.GetSessionFromContext(r)
	product, err := c.productStore.GetProduct(uint(productID))
	if err != nil || product == nil || product.TenantID != sess.TenantID {
		ShowToast(w, "Produto nao encontrado", "error")
		return
	}

	cart, err := c.resolveOpenCart(r, w, sess)
	if err != nil {
		ShowToast(w, "Erro ao preparar carrinho", "error")
		return
	}

	if err := c.cartStore.AddOrIncrementItem(cart.ID, product.ID, qty, product.CostPrice); err != nil {
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

func (c *ClientHandler) resolveOpenCart(r *http.Request, w http.ResponseWriter, sess *store.Session) (*store.Cart, error) {
	if sess.CartID != 0 {
		cart, err := c.cartStore.FindOpenByID(sess.CartID, sess.TenantID, sess.ContactInfo.ID)
		if err == nil {
			return cart, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	fmt.Println(sess.ContactInfo.ID)

	cart, err := c.cartStore.FindOpenByContact(sess.TenantID, sess.ContactInfo.ID)
	if err == nil {
		if setErr := c.sessionStore.SetCartID(r, w, cart.ID); setErr != nil {
			return nil, setErr
		}
		return cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

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

func (c *ClientHandler) PostConfirmOrder(w http.ResponseWriter, r *http.Request) {
	sess := middleware.GetSessionFromContext(r)

	var cart *store.Cart
	var err error

	if sess.CartID != 0 {
		cart, err = c.cartStore.FindOpenByID(sess.CartID, sess.TenantID, sess.ContactInfo.ID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			ShowToast(w, "Erro ao preparar carrinho", "error")
			return
		}
	}

	if cart == nil {
		cart, err = c.cartStore.FindOpenByContact(sess.TenantID, sess.ContactInfo.ID)
		if err != nil {
			ShowToast(w, "Carrinho vazio", "error")
			return
		}
		_ = c.sessionStore.SetCartID(r, w, cart.ID)
	}

	_, err = c.orderStore.ConfirmFromCart(cart.ID, sess.TenantID, sess.ContactInfo.ID)
	if err != nil {
		ShowToast(w, "Erro ao confirmar pedido", "error")
		return
	}

	_ = c.sessionStore.SetCartID(r, w, 0)
	w.Header().Set(HXRedirect, "/client/items")
	w.WriteHeader(http.StatusOK)
}
