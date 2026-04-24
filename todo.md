1. Secret de sessão hardcoded — crítico
  // main.go linha 48 e 53
  "VERYSECRETKEY" // TODO: Colocar no env
  Qualquer um que leia o repositório pode forjar cookies de sessão e se autenticar como qualquer usuário. Precisa ir para variável de ambiente antes de qualquer deploy.

  2. fmt.Println em produção
  Tem 5 fmt.Println espalhados em handlers e no session store — vaza dados de sessão e erros internos no stdout sem estrutura. Troca por log/slog ou similar.

  3. PerPage: 1000 no handler de novo pedido
  // adminOrder.go linha 67
  Pagination: store.Pagination{Page: 1, PerPage: 1000}
  O comentário // TODO: No futuro vai dar ruim já diz tudo. Você já resolveu para produtos — contatos precisa do mesmo tratamento (busca com HTMX).

  4. Zero testes
  Nenhum arquivo _test.go no projeto. Lógica crítica como cálculo de preço, confirmação de pedido e validação de formulário não tem cobertura alguma.

  ---
  O que está bem feito

  - Multi-tenancy correto — todo query filtra por TenantID, sem risco de vazamento entre tenants
  - Arquitetura de interfaces — handlers dependem de interfaces (ProductStore, CartStore), não de implementações concretas. Fácil de testar e trocar banco
  - Sessão separada por role — admin e client têm cookies distintos, sem mistura de contexto
  - Snapshot de pedido — ao confirmar, os itens do carrinho viram OrderItem com preço no momento da compra. Correto
  - Graceful shutdown em main.go
  - Paginação com filtros — FindAllByUserWithFilters é reutilizado em vários contextos

  ---
  Vale refatorar antes de produção (não bloqueia, mas importa)

  - cookiesotore — typo no nome do pacote. Menor, mas vai incomodar sempre
  - FindAllByUser sem limite — ainda existe na interface, só FindAllByUserWithFilters deveria existir
  - Erros retornados como http.Error em alguns handlers, ShowToast em outros — inconsistente. Em HTMX, http.Error não aparece pro usuário de forma útil

  ---
  Resumo: a fundação é sólida. Resolva o secret de sessão, tire os fmt.Println e adicione alguns testes nas partes críticas — aí está em forma para produção.
