package services

import (
	"errors"

	"github.com/DTineli/ez/internal/store"
)

var ErrPriceTableHasContacts = errors.New("tabela possui clientes vinculados")

type PriceTableService interface {
	Create(
		tenantID uint,
		name string,
		percentage float64,
	) (*store.PriceTable, error)
	Delete(id, tenantID uint) error
	FindAll(tenantID uint) ([]store.PriceTable, error)
	FindAllActive(tenantID uint) ([]store.PriceTable, error)
	FindAllActiveByContact(tenantID, contactID uint) ([]store.PriceTable, error)
	FindOne(id, tenantID uint) (*store.PriceTable, error)
	GetOne(id, tenantID uint) (*store.PriceTable, error)
	Apply(costPrice float64, pt *store.PriceTable) float64

	AddPrice(tableID, variationID uint, price float64) (uint, error)
	GetProductPrice(id uint) (*store.ProductPrice, error)
	UpdatePrice(id, tenantID uint, price float64) error
	RemovePrice(priceID, tenantID uint) error
	SearchVariants(
		tenantID, priceTableID uint,
		q string,
	) ([]store.Variant, error)
}

type priceTableService struct {
	store store.PriceTableStore
}

func NewPriceTableService(s store.PriceTableStore) PriceTableService {
	return &priceTableService{store: s}
}

func (p *priceTableService) AddPrice(
	tableID, variationID uint,
	price float64,
) (uint, error) {
	pPrice := store.ProductPrice{
		Price:        price,
		VariantID:    variationID,
		PriceTableID: tableID,
	}
	if err := p.store.CreateProductPrice(&pPrice); err != nil {
		return 0, err
	}
	return pPrice.ID, nil
}

type PriceProductInfo struct {
}

func (p *priceTableService) GetAllPriceTablesInfoByProduct(
	id uint,
) (*PriceProductInfo, error) {

	return &PriceProductInfo{}, nil
}

func (p *priceTableService) GetProductPrice(
	id uint,
) (*store.ProductPrice, error) {
	return p.store.GetOneProductPriceWithVariant(id)
}

func (p *priceTableService) RemovePrice(priceID, tenantID uint) error {
	pp, err := p.store.GetOneProductPrice(priceID)
	if err != nil {
		return err
	}
	if _, err := p.store.GetOne(pp.PriceTableID, tenantID); err != nil {
		return err
	}
	return p.store.DeleteProductPrice(priceID)
}

func (p *priceTableService) UpdatePrice(
	id, tenantID uint,
	price float64,
) error {
	pp, err := p.store.GetOneProductPrice(id)
	if err != nil {
		return err
	}
	if _, err := p.store.GetOne(pp.PriceTableID, tenantID); err != nil {
		return err
	}
	return p.store.UpdateProductPrice(id, price)
}

func (p *priceTableService) Create(
	tenantID uint,
	name string,
	percentage float64,
) (*store.PriceTable, error) {
	table := &store.PriceTable{
		Name:       name,
		Percentage: percentage,
		TenantID:   tenantID,
	}
	if err := p.store.CreatePriceTable(table); err != nil {
		return nil, err
	}
	return table, nil
}

func (p *priceTableService) Delete(id, tenantID uint) error {
	has, err := p.store.HasContacts(id, tenantID)
	if err != nil {
		return err
	}
	if has {
		return ErrPriceTableHasContacts
	}
	return p.store.Delete(id, tenantID)
}

func (p *priceTableService) FindAll(tenantID uint) ([]store.PriceTable, error) {
	return p.store.FindAllByTenant(tenantID)
}

func (p *priceTableService) FindAllActive(
	tenantID uint,
) ([]store.PriceTable, error) {
	return p.store.FindAllActiveByTenant(tenantID)
}

func (p *priceTableService) FindAllActiveByContact(
	tenantID, contactID uint,
) ([]store.PriceTable, error) {
	return p.store.FindAllActiveByTenantAndClient(tenantID, contactID)
}

func (p *priceTableService) FindOne(
	id, tenantID uint,
) (*store.PriceTable, error) {
	return p.store.GetOneWithPrices(id, tenantID)
}

func (p *priceTableService) GetOne(
	id, tenantID uint,
) (*store.PriceTable, error) {
	return p.store.GetOne(id, tenantID)
}

func (p *priceTableService) Apply(
	costPrice float64,
	pt *store.PriceTable,
) float64 {
	return ApplyPriceTable(costPrice, pt)
}

func (p *priceTableService) SearchVariants(
	tenantID, priceTableID uint,
	q string,
) ([]store.Variant, error) {
	return p.store.SearchVariantsForPriceTable(tenantID, priceTableID, q)
}

// ApplyPriceTable aplica o multiplicador da tabela ao custo base.
// Retorna costPrice sem alteração se pt for nil.
func ApplyPriceTable(costPrice float64, pt *store.PriceTable) float64 {
	if pt == nil {
		return costPrice
	}
	return costPrice * (1 + pt.Percentage/100)
}
