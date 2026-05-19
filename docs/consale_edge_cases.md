# Consale — Edge Cases: Relacionamento Cliente e Loja

Use este documento para ir respondendo cada caso. Adicione sua resposta abaixo de cada item.

---

## 1. Acesso via URL

### 1.1 Acesso sem slug (`consale.com`)
Usuário entra na raiz sem nenhum slug de admin.

- [X] Mostra landing page genérica
- [ ] Retorna 404
- [ ] Redireciona pro último slug acessado (cookie/sessão)
- [ ] Outro: ___

**Resposta:**

Landing page com todos as loja q ele tem acesso 

---

### 1.2 Slug inválido ou inexistente (`xyz.consale.com`)
O slug não corresponde a nenhum admin cadastrado.

- [ ] Página de erro 404 customizada
- [X] Página genérica "loja não encontrada"
- [ ] Redireciona para `consale.com`
- [ ] Outro: ___

**Resposta:**

Pagina de Loja nao encontrada com botao de redirect para conssale.com

---

### 1.3 Admin desativado mas slug ainda acessível
O admin foi suspenso/desativado, mas a URL ainda é válida tecnicamente.

- [ ] Bloqueia acesso, mostra mensagem genérica
- [X] Redireciona para `consale.com`
- [ ] Mantém acesso (só novas compras bloqueadas)
- [ ] Outro: ___

**Resposta:**

Redireciona para consale

---

## 2. Autenticação e Sessão

### 2.1 Cliente já logado entra num link de outro admin
User tem sessão ativa e acessa o slug de um admin diferente do que usou pra logar.

- [X] Acessa direto sem fricção (sessão global vale pra qualquer slug)
- [ ] Pede confirmação ("você está acessando como X, continuar?")
- [ ] Força novo login
- [ ] Outro: ___

**Resposta:**

---

### 2.2 Cliente logado mas sem customer nesse admin
User tem conta, acessou o slug, mas o admin ainda não cadastrou um customer para ele.

- [X] Bloqueia — mostra "aguarde liberação do administrador"
- [ ] Permite navegar mas bloqueia na hora de comprar
- [ ] Cria o vínculo automaticamente (customer gerado on-the-fly)
- [ ] Outro: ___

**Resposta:**

---

### 2.3 Sessão expirada no meio do carrinho
Cliente estava com itens no carrinho, sessão expirou, fez login novamente.

- [ ] Carrinho é perdido, começa do zero
- [X] Carrinho é recuperado (salvo no servidor por customer)
- [ ] Carrinho é recuperado por cookie/localStorage
- [ ] Outro: ___

**Resposta:**

---

## 3. Customer e Vínculo

### 3.1 Admin cadastra customer com e-mail de user já existente
O e-mail informado já tem uma conta global (user de outro admin).

- [X] Vincula silenciosamente ao user existente
- [ ] Avisa o admin antes de vincular
- [ ] Cria um user novo separado
- [ ] Outro: ___

**Resposta:**

---

### 3.2 Admin remove um customer
O customer é deletado ou desativado pelo admin.

- [X] User perde acesso imediatamente
- [ ] User perde acesso apenas no próximo login
- [X] Pedidos em aberto continuam visíveis pro admin mas o cliente não acessa mais
- [ ] O que acontece com pedidos pendentes? 
- [ ] Outro: ___

**Resposta:**

Pedidos pendentes ficam na pagina de pedidos do admin e na paginal geral do User

---

### 3.3 Mesmo user tenta ter dois customers no mesmo admin
Cenário de inconsistência: por algum bug ou race condition, o mesmo user fica vinculado a dois customers do mesmo admin.

- [X] Impedido no banco (constraint única user+admin)
- [ ] Sistema usa o mais recente
- [ ] Sistema usa o mais antigo
- [ ] Outro: ___

**Resposta:**

---

## 4. Tabelas de Preço

### 4.1 Customer sem nenhuma tabela atribuída
O admin criou o customer mas esqueceu de atribuir tabela de preço.

- [X] Loja aparece vazia (sem produtos)
- [X] Mensagem explicativa ("nenhum produto disponível no momento")
- [ ] Bloqueia o acesso à loja até ter tabela
- [ ] Outro: ___

**Resposta:**

---

### 4.2 Admin remove uma tabela já atribuída a customers ativos
A tabela é deletada ou desvinculada enquanto customers a usavam.

- [X] Produtos daquela tabela somem imediatamente do catálogo
- [ ] Produtos somem só após cache expirar
- [ ] Carrinho com itens dessa tabela é afetado — como? ___
- [ ] Pedidos já feitos com essa tabela são afetados — como? ___
- [ ] Outro: ___

**Resposta:**

Pedidos ja feitos nao sao afetados, somente novos pedidos nao podem ser criados
com aquela tabela

---

### 4.3 Produto em duas tabelas do mesmo customer com preços diferentes
Customer tem duas tabelas atribuídas e o mesmo produto aparece nas duas com valores distintos.

- [ ] Prevalece o menor preço
- [ ] Prevalece o maior preço
- [ ] Prevalece a tabela de maior prioridade (ordem de atribuição)
- [ ] Mostra as duas opções pro cliente
- [ ] Outro: ___

**Resposta:**

Todas as tabelas tem todos os produtos

---

## 5. Carrinho e Pedido

### 5.1 Preço alterado enquanto item está no carrinho
Admin muda o preço na tabela depois que o cliente já adicionou o produto ao carrinho.

- [ ] Preço congela no momento da adição (snapshot)
- [X] Preço atualiza automaticamente no carrinho
- [ ] Avisa o cliente na hora de finalizar ("preço atualizado")
- [ ] Outro: ___

**Resposta:**

---

### 5.2 Pedido aguardando aprovação por muito tempo
Admin demora dias para aprovar ou rejeitar.

- [X] Pedido não expira, fica aberto indefinidamente
- [ ] Expira após N dias — quantos? ___
- [X] Cliente pode cancelar manualmente enquanto aguarda
- [ ] Outro: ___

**Resposta:**

---

### 5.3 Admin rejeita pedido sem dar motivo
Pedido rejeitado sem justificativa.

- [X] Cliente vê apenas "pedido rejeitado"
- [ ] Campo de motivo é obrigatório para o admin
- [X] Campo de motivo é opcional mas exibido ao cliente se preenchido
- [ ] Outro: ___

**Resposta:**

---

### 5.4 Cliente faz dois pedidos simultâneos
Cliente tenta abrir um segundo pedido antes do primeiro ser resolvido.

- [X] Permitido — múltiplos pedidos em aberto ao mesmo tempo
- [ ] Bloqueado — precisa aguardar resolução do pedido anterior
- [ ] Limitado — pode ter até N pedidos abertos
- [ ] Outro: ___

**Resposta:**

---

## Resumo de pendências

| # | Edge case | Resolvido? |
|---|---|---|
| 1.1 | Acesso sem slug | |
| 1.2 | Slug inválido | |
| 1.3 | Admin desativado | |
| 2.1 | Logado, entra em outro admin | |
| 2.2 | Logado, sem customer no admin | |
| 2.3 | Sessão expirada no carrinho | |
| 3.1 | E-mail já existente no cadastro | |
| 3.2 | Admin remove customer | |
| 3.3 | User duplicado no mesmo admin | |
| 4.1 | Customer sem tabela | |
| 4.2 | Tabela removida com customers ativos | |
| 4.3 | Produto em duas tabelas com preços diferentes | |
| 5.1 | Preço alterado com item no carrinho | |
| 5.2 | Pedido sem aprovação por muito tempo | |
| 5.3 | Rejeição sem motivo | |
| 5.4 | Dois pedidos simultâneos | |
