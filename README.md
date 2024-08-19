## Pré-requisitos

Antes de começar, certifique-se de ter os seguintes itens instalados:

- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Make](https://www.gnu.org/software/make/)

## Como Rodar o Projeto

### 1. Clonar o Repositório

Primeiro, clone o repositório para a sua máquina local:

```sh
git clone https://github.com/alissoncorsair/appsolidario-backend.git
cd appsolidario-backend 
docker compose up -d (para rodar o app + banco)
make migrate-up-docker (para rodar as migrations e criar as tabelas)
```

### 2. Setup do ambiente

Copie o arquivo `.env.example` para `.env` e preencha com as variáveis de ambiente necessárias.

```sh
cp .env.example .env
```

### 3. Iniciar o projeto

Execute o comando abaixo para iniciar o projeto:

```sh
make docker-up
```

### 4. Rodar as migrations

Após rodar o comando acima, rode o comando abaixo para rodar as migrations:

```sh
make docker-migrate-up
```