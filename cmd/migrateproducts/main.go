// Comando one-off: migra produtos do bancoProd.db (SQLite, schema antigo)
// pro Postgres de produção (schema atual), via Product+Variant+Attribute.
//
// Uso:
//
//	set -a && . ./.env && set +a && go run ./cmd/migrateproducts --dry-run
//	set -a && . ./.env && set +a && go run ./cmd/migrateproducts
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/DTineli/ez/cmd/migrateproducts/legacy"
	"github.com/DTineli/ez/cmd/migrateproducts/migration"
	"github.com/DTineli/ez/internal/config"
	database "github.com/DTineli/ez/internal/store/db"
	"github.com/DTineli/ez/internal/store/dbstore"

	_ "modernc.org/sqlite"
)

func main() {
	sqlitePath := flag.String("sqlite-path", "./bancoProd.db", "caminho do bancoProd.db")
	dryRun := flag.Bool("dry-run", false, "simula a migração sem escrever no Postgres")
	limit := flag.Int("limit", 0, "limita o número de linhas processadas (0 = sem limite, pra testes)")
	flag.Parse()

	cfg := config.MustLoadConfig()
	pg := database.MustOpen(cfg.ActiveDatabaseURL())

	sqliteDB, err := sql.Open("sqlite", *sqlitePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro abrindo %s: %v\n", *sqlitePath, err)
		os.Exit(1)
	}
	defer sqliteDB.Close()

	products, err := legacy.ReadAll(sqliteDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo products de %s: %v\n", *sqlitePath, err)
		os.Exit(1)
	}
	if *limit > 0 && len(products) > *limit {
		products = products[:*limit]
	}

	variants, err := legacy.ReadVariants(sqliteDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo variants de %s: %v\n", *sqlitePath, err)
		os.Exit(1)
	}
	attrs, err := legacy.ReadAttributes(sqliteDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo attributes de %s: %v\n", *sqlitePath, err)
		os.Exit(1)
	}
	attrValues, err := legacy.ReadAttributeValues(sqliteDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo attribute_values de %s: %v\n", *sqlitePath, err)
		os.Exit(1)
	}
	variantAttrs, err := legacy.ReadVariantAttributes(sqliteDB)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro lendo variant_attributes de %s: %v\n", *sqlitePath, err)
		os.Exit(1)
	}

	mapped, mapWarnings := migration.MapAll(products, variants)

	tenantStore := dbstore.NewTenantStore(pg)
	productStore := dbstore.NewProductStore(pg)

	missing, err := migration.ValidateTenants(mapped, tenantStore)
	if err != nil {
		fmt.Fprintf(os.Stderr, "erro validando tenants: %v\n", err)
		os.Exit(1)
	}
	if len(missing) > 0 {
		fmt.Fprintf(os.Stderr, "ABORTADO: tenant(s) %v não existem no Postgres de destino. Nada foi escrito.\n", missing)
		os.Exit(1)
	}

	summary, oldVariantIDToNew := migration.Run(pg, productStore, mapped, *dryRun)
	summary.Warnings = append(summary.Warnings, mapWarnings...)

	linked, linkWarnings := migration.LinkAttributes(productStore, attrs, attrValues, variantAttrs, oldVariantIDToNew, *dryRun)
	summary.AttributesLinked = linked
	summary.Warnings = append(summary.Warnings, linkWarnings...)

	printSummary(*dryRun, summary)

	if summary.Skipped > 0 {
		os.Exit(2)
	}
}

func printSummary(dryRun bool, s migration.Summary) {
	if dryRun {
		fmt.Println("=== DRY RUN — nada foi escrito no Postgres ===")
	}
	fmt.Printf("Linhas lidas (já filtrado tenant ignorado): %d\n", s.RowsRead)
	fmt.Printf("Produtos criados: %d\n", s.ProductsCreated)
	fmt.Printf("Variantes criadas: %d\n", s.VariantsCreated)
	fmt.Printf("Variantes corrigidas (já existiam, dados zerados): %d\n", s.VariantsUpdated)
	fmt.Printf("Atributos linkados (variant_attributes): %d\n", s.AttributesLinked)
	fmt.Printf("Pulados: %d\n", s.Skipped)

	if len(s.SkippedDetails) > 0 {
		fmt.Println("\n--- Detalhes dos pulados ---")
		for _, d := range s.SkippedDetails {
			fmt.Println(" -", d)
		}
	}

	if len(s.Warnings) > 0 {
		fmt.Println("\n--- Avisos (campos descartados / defaults aplicados) ---")
		for _, w := range s.Warnings {
			fmt.Println(" -", w)
		}
	}
}
