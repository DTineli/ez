# Migração: Supabase Free → PostgreSQL no VPS

> **Por que migrar agora?** Você ainda não tem usuários — momento ideal, zero risco. O Supabase Free pausa o projeto após 7 dias sem atividade, causando cold starts de 5–15s. Com Postgres no mesmo VPS da aplicação, a latência cai para milissegundos (conexão localhost).

---

## Comparativo

| Critério | Supabase Free | Postgres no VPS |
|---|---|---|
| Latência | Alta (cold start 5–15s) | Mínima (localhost) |
| Custo | Grátis (limitado) | Já incluso no VPS |
| Controle | Nenhum | Total |
| Backups | Automático | Você configura (cron) |
| Cold start | Sim (pausa em 7 dias) | Não |
| Escalabilidade | Limitado no free | Depende do VPS |
| Setup inicial | Zero | ~30–60 minutos |

---

## Pré-requisitos

- Acesso SSH ao VPS (Ubuntu 22.04 / Debian recomendado)
- Acesso ao painel do Supabase para exportar os dados
- `pg_dump` instalado na sua máquina local
- Variáveis de ambiente configuráveis na aplicação

---

## Passo 1 — Exportar dados do Supabase

No painel do Supabase, vá em **Settings → Database** e copie a connection string. Depois, no seu terminal local:

```bash
# Exportar schema e dados completos
pg_dump "postgres://USER:PASSWORD@db.PROJETO.supabase.co:5432/postgres" \
  --no-owner --no-acl \
  -f backup_supabase.sql

# Verificar se o arquivo foi gerado
wc -l backup_supabase.sql
```

> Use `--no-owner` e `--no-acl` para evitar erros de permissão ao importar em outro servidor.

---

## Passo 2 — Instalar PostgreSQL no VPS

Conecte via SSH no seu VPS:

```bash
# Atualizar pacotes
sudo apt update && sudo apt upgrade -y

# Instalar PostgreSQL
sudo apt install postgresql postgresql-contrib -y

# Habilitar e iniciar o serviço
sudo systemctl enable postgresql
sudo systemctl start postgresql

# Verificar se está rodando
sudo systemctl status postgresql
```

---

## Passo 3 — Criar usuário e banco de dados

```bash
# Entrar como superusuário postgres
sudo -u postgres psql
```

```sql
-- Dentro do psql:
CREATE USER appuser WITH PASSWORD 'senha_forte_aqui';
CREATE DATABASE appdb OWNER appuser;
GRANT ALL PRIVILEGES ON DATABASE appdb TO appuser;
\q
```

---

## Passo 4 — Importar os dados

```bash
# Copiar o backup para o VPS (rode no seu terminal local)
scp backup_supabase.sql usuario@IP_DO_VPS:/home/usuario/

# No VPS: importar o banco
psql -U appuser -d appdb -f /home/usuario/backup_supabase.sql

# Verificar se as tabelas foram importadas
psql -U appuser -d appdb -c "\dt"
```

---

## Passo 5 — Configurar pg_hba.conf

Por padrão o Postgres já aceita conexões locais, mas vale confirmar:

```bash
# Encontrar o arquivo
sudo -u postgres psql -c "SHOW hba_file;"

# Abrir para editar (ajuste o caminho conforme a saída acima)
sudo nano /etc/postgresql/14/main/pg_hba.conf
```

Garanta que essas linhas existem:

```
local   all   all                  scram-sha-256
host    all   all   127.0.0.1/32   scram-sha-256
```

```bash
# Reiniciar após editar
sudo systemctl restart postgresql
```

---

## Passo 6 — Atualizar a connection string no Go

```bash
# Antes (Supabase)
DATABASE_URL=postgres://USER:PASS@db.PROJETO.supabase.co:5432/postgres

# Depois (Postgres local)
DATABASE_URL=postgres://appuser:senha_forte_aqui@localhost:5432/appdb?sslmode=disable
```

No código Go, nenhuma mudança necessária — só a connection string:

```go
import "github.com/jackc/pgx/v5"

connStr := os.Getenv("DATABASE_URL")
conn, err := pgx.Connect(context.Background(), connStr)
if err != nil {
    log.Fatal("Erro ao conectar:", err)
}
defer conn.Close(context.Background())
```

---

## Passo 7 — Testar a conexão

```bash
# Testar direto no terminal do VPS
psql -U appuser -h localhost -d appdb

# Dentro do psql
SELECT COUNT(*) FROM nome_da_sua_tabela;
\q

# Subir a aplicação e verificar
go run main.go
```

---

## Backups automáticos

Crie um script de backup:

```bash
sudo nano /home/usuario/backup_db.sh
```

```bash
#!/bin/bash
DATE=$(date +%Y-%m-%d_%H-%M)
BACKUP_DIR="/home/usuario/backups"
mkdir -p $BACKUP_DIR

pg_dump -U appuser appdb > $BACKUP_DIR/backup_$DATE.sql

# Manter apenas os últimos 7 backups
ls -t $BACKUP_DIR/backup_*.sql | tail -n +8 | xargs rm -f
```

```bash
# Dar permissão de execução
chmod +x /home/usuario/backup_db.sh

# Configurar cron para rodar todo dia às 3h
crontab -e

# Adicionar a linha:
0 3 * * * /home/usuario/backup_db.sh
```

---

## Segurança

**Nunca exponha a porta 5432 para a internet.** Configure o firewall:

```bash
sudo ufw allow ssh
sudo ufw allow 80
sudo ufw allow 443
sudo ufw deny 5432
sudo ufw enable
```

Se precisar acessar o banco remotamente, use SSH tunnel:

```bash
# Rode no seu computador local
ssh -L 5432:localhost:5432 usuario@IP_DO_VPS

# Depois conecte normalmente em localhost:5432 com seu cliente (DBeaver, etc.)
```

---

## Quando voltar para um serviço gerenciado?

Considere Supabase Pro, Railway, Neon ou RDS quando:

- Precisar de alta disponibilidade e failover automático
- Volume de dados crescer além da capacidade do VPS
- Custo do downtime superar o custo do serviço (~$25/mês)
- Time sem capacidade de gerenciar infra de banco

**Por enquanto, Postgres local é a escolha certa:** zero custo adicional, latência de localhost, controle total.
