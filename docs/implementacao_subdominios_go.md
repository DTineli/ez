# 🌐 Implementação de Subdomínios (Multi-tenant) em Go

Guia prático para implementar `cliente.ez.com` dentro da aplicação.

------------------------------------------------------------------------

# 🎯 Objetivo

Permitir que cada cliente tenha acesso ao sistema através de um
subdomínio próprio:

    lojadojoao.ez.com
    lojademaria.ez.com
    empresa123.ez.com

A aplicação deve:

-   Identificar o subdomínio automaticamente
-   Descobrir qual cliente está acessando
-   Carregar dados, pedidos e configurações daquele cliente
-   Isolar os dados entre clientes

------------------------------------------------------------------------

# 🧱 Arquitetura Geral

Fluxo completo:

    1. Usuário acessa lojadojoao.ez.com
    2. DNS resolve para o servidor
    3. Go recebe a request
    4. Middleware extrai subdomínio
    5. Sistema identifica tenant (cliente)
    6. Dados do cliente são carregados
    7. Request segue normalmente

------------------------------------------------------------------------

# 🌐 Etapa 1 --- Configurar DNS

Criar um wildcard para aceitar qualquer subdomínio.

    *.ez.com → IP do servidor

Sem isso, os subdomínios não chegam até sua aplicação.

------------------------------------------------------------------------

# ⚙️ Etapa 2 --- Proxy reverso (recomendado)

Usar Nginx para redirecionar todos subdomínios para o Go.

Exemplo conceitual:

    server_name *.ez.com;
    proxy_pass http://localhost:8080;

Isso evita precisar configurar cada cliente manualmente.

------------------------------------------------------------------------

# 🧩 Etapa 3 --- Middleware para extrair subdomínio

Criar middleware responsável por:

-   ler o host
-   extrair subdomínio
-   colocar no contexto da request

Exemplo de lógica:

    host: lojadojoao.ez.com
    subdomain: lojadojoao

Regras:

-   ignorar www
-   ignorar domínio raiz
-   validar formato

------------------------------------------------------------------------

# 🧠 Etapa 4 --- Identificação do Tenant

Criar tabela de clientes:

    clients
    - id
    - name
    - slug
    - active
    - created_at

Exemplo:

    name: Loja do João
    slug: lojadojoao

O slug será o subdomínio.

------------------------------------------------------------------------

# 🔎 Etapa 5 --- Resolver cliente por subdomínio

Fluxo interno:

    1. middleware pega subdomain
    2. busca cliente no banco:

       SELECT * FROM clients WHERE slug = ?

    3. valida:
       - existe?
       - ativo?
       - plano válido?

    4. salva cliente no context

Depois disso, toda request sabe quem é o tenant.

------------------------------------------------------------------------

# 🧾 Etapa 6 --- Estrutura de dados multi-tenant

## Opção recomendada (MVP + produção inicial)

Banco único com coluna tenant_id:

    orders
    - id
    - tenant_id
    - customer_name
    - total
    - created_at

Toda query deve filtrar:

    WHERE tenant_id = ?

------------------------------------------------------------------------

# 🧱 Etapa 7 --- Camada de acesso a dados

Criar padrão:

    GetOrdersByTenant(tenantID)
    CreateOrder(tenantID)
    GetCustomersByTenant(tenantID)

Nunca buscar dados globais sem tenant.

------------------------------------------------------------------------

# 🔐 Etapa 8 --- Segurança obrigatória

Nunca confiar apenas no subdomínio.

Sempre validar:

-   cliente existe
-   cliente ativo
-   não bloqueado
-   plano válido

Se falhar:

    404 ou 403

------------------------------------------------------------------------

# 🎨 Etapa 9 --- Customização por cliente

Possibilidades futuras:

-   logo próprio
-   cores
-   layout
-   domínio próprio
-   regras específicas

Adicionar campos em clients:

    theme_color
    logo_url
    custom_domain

------------------------------------------------------------------------

# 🌍 Etapa 10 --- Domínio raiz (ez.com)

Decidir função do domínio principal:

Opções:

-   landing page
-   página institucional
-   login administrativo
-   onboarding de clientes

------------------------------------------------------------------------

# 🚀 Etapa 11 --- Onboarding automático

Fluxo ideal:

    1. novo cliente se cadastra
    2. sistema gera slug automático
    3. salva em clients
    4. subdomínio passa a funcionar automaticamente

Exemplo:

    Nome: Clínica Vida
    Slug gerado: clinicavida

    → clinicavida.ez.com

------------------------------------------------------------------------

# 🧪 Etapa 12 --- Testes necessários

Testar:

-   acesso com subdomínio válido
-   acesso com subdomínio inexistente
-   domínio raiz
-   www
-   cliente desativado
-   cliente ativo

------------------------------------------------------------------------

# 📦 Etapa 13 --- Estrutura de pastas sugerida

    /internal
      /middleware
        tenant.go

      /services
        tenant_service.go

      /repositories
        tenant_repository.go

      /handlers
        orders_handler.go

------------------------------------------------------------------------

# 🔮 Evoluções futuras

## Domínio próprio do cliente

Permitir:

    lojadojoao.com → CNAME → ez.com

Sistema precisa:

-   reconhecer domínio
-   mapear para tenant

------------------------------------------------------------------------

## Banco por cliente (escala grande)

Hoje:

    1 banco

Futuro:

    1 banco por tenant

------------------------------------------------------------------------

## Cache por tenant

Redis:

    tenant:lojadojoao:orders

------------------------------------------------------------------------

# 🧠 Regras de ouro

1.  Toda request precisa saber quem é o tenant
2.  Nunca consultar dados sem tenant_id
3.  Subdomínio é identidade do cliente
4.  Middleware resolve 80% da complexidade
5.  Modelagem inicial define escalabilidade futura

------------------------------------------------------------------------

# 📌 Checklist de implementação

## Infra

-   [ ] wildcard DNS
-   [ ] proxy reverso configurado

## Backend

-   [ ] middleware de subdomínio
-   [ ] tabela clients
-   [ ] service de tenant
-   [ ] resolver tenant por slug
-   [ ] salvar tenant no context

## Dados

-   [ ] adicionar tenant_id nas tabelas
-   [ ] ajustar queries

## Segurança

-   [ ] validação de cliente ativo
-   [ ] fallback para erro

## Produto

-   [ ] fluxo de criação de cliente
-   [ ] geração automática de slug

------------------------------------------------------------------------

# 🧭 Estado final esperado

Sistema funcionando assim:

    cliente cria conta
    ↓
    slug gerado
    ↓
    cliente.ez.com começa a responder
    ↓
    dados isolados
    ↓
    mesmo backend para todos
    ↓
    multi-tenant ativo

------------------------------------------------------------------------

# 📎 Próxima etapa

Depois de implementar isso:

1.  autenticação por tenant
2.  permissões por usuário
3.  painel admin global
4.  billing por cliente
5.  métricas por tenant
