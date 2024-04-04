# CINEMA GRADER

Cinema Grader é uma API RESTful feita como forma de treinar a criação de servers JSON em Go. 
A API é construída usando a Standard Library de Go, o framework `Fiber` e o banco de dados `PosgreSQL`. Além disso,
existem algumas libraries auxiliares para auxiliar com autenticação (`bcrypt`, `jwt`) e outras para configurações e facilitamento do desenvolvimento
(`godotenv`, `air`).

É uma API que simula uma versão simplificada de uma aplicação como o IMDB. Nessa API é possível
- Em usuários:
  - Comuns
    - Fazer um CRUD de usuário (exclusivamente a si mesmo)
    - Visualizar filmes criados na API
    - Visualizar informação sobre atores
    - Visualizar que filmes um certo ator participou
    - Visualizar quais atores participaram de um certo filme
    - CRUD de comentários em filmes (e possibilidade de dar notas em filmes)
  - Admins
    - CRUD de filmes na API
    - CRUD de atores na API
    - CRUD de usuários (tanto a si mesmo quanto outros)
    - CRUD de comentários em filmes (de qualquer usuário)
    - Associar filmes com atores e vice versa
- Como já notado pelos items acima, a API integra autenticação e utiliza diversos middlewares. Consulte a seção de documentação neste README para ver mais detalhes sobre rotas.

## Rodando a API
1. Instale a ferramenta (Air)[https://github.com/cosmtrek/air]. Se você tem acesso a minha playlist de backend, tem detalhes sobre como fazer isso na Aula número **4** da playlist de Backend.
2. Rode um banco **PostgreSQL** na versão **16.2**, seja localmente ou remotamente (novamente, se tiver acesso e quiser saber como rodar um banco PostgreSQL localmente, aula **5** da minha playlist de backend).
3. Crie um arquivo `.env` na raiz do repo, copiado do arquivo `.env.example`.
4. Preencha as variáveis de ambiente conforme o que está comentado no arquivo.
5. Se você tiver o `make` instalado, pode rodar `make` no terminal que ele vai executar o `Makefile` na raiz do projeto.
6. Alternativamente, rode o comando `air` em um terminal na raiz do repo. Pronto! Agora o código está rodando e você pode executar suas requisições a vontade.

## Documentação
Na pasta `api` na raiz do diretório temos
1. Um arquivo `c_grader.json` que é um arquivo de configuração do API Client (Bruno)[https://www.usebruno.com/] que mostra todas as rotas com requisições já prontas para elas
2. O DER (Diagrama de entidades e relações) da nossa database, que demonstra quais as tabelas que existem e a relação entre elas. Foi feita no site (Draw.io)[draw.io] e recomendo abrir a imagem dentro do arquivo para facilitar a leitura, visto que o png tem alguns defeitos de visualização.