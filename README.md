# CINEMA GRADER

Cinema Grader é uma API RESTful feita como forma de treinar a criação de servers JSON em Go. 
A API é construída usando a Standard Library de Go, o framework `Fiber` e o banco de dados `PosgreSQL`. Além disso, existem algumas libraries auxiliares para auxiliar com autenticação (`bcrypt`, `jwt`) e outras para configurações e facilitamento do desenvolvimento
(`godotenv`, `air`).

É uma API que simula uma versão simplificada de uma aplicação como o IMDB, e possui interações com diversos tipos relacionamentos, sendo possível criar usuários, filmes, atores e comentários. Consulte a sessão de documentação para ver todas as rotas e possibilidades da api.

A api possui testes automatizados de integração para todas as rotas e testes unitários onde foi julgado como necessário.

## Rodando a API
1. Instale a ferramenta [Air](https://github.com/cosmtrek/air). Se você tem acesso a minha playlist de backend, tem detalhes sobre como fazer isso na Aula número **4** da playlist de Backend.
2. Rode um banco **PostgreSQL** na versão **16.2**, seja localmente ou remotamente (novamente, se tiver acesso e quiser saber como rodar um banco PostgreSQL localmente, aula **5** da minha playlist de backend no youtube).
3. Crie um arquivo `.env` na raiz do repo, copiado do arquivo `.env.example`.
4. Preencha as variáveis de ambiente conforme o que está comentado no arquivo.
5. Se você tiver o `make` instalado, pode rodar `make` no terminal que ele vai executar o `Makefile` na raiz do projeto.
6. Alternativamente, rode o comando `air` em um terminal na raiz do repo. Pronto! Agora o código está rodando e você pode executar suas requisições a vontade.
7. Essa API possui testes automatizados. Para rodá-los, execute o comando `make test` (ou `go test ./...`) na raiz do projeto, que irá recursivamente consultar todas as pastas do repositório e rodar os testes encontrados. Caso queira rodar alguma pasta específica, é só colocar o caminho dela como argumento ao invés do `./...` (ex: `go test ./tests`). Testes de integração estão na pasta `tests` e os testes unitários estão na mesma pasta que seus arquivos, como dita o paradigma de testes automatizados da linguagem.

## Documentação
Na pasta `api` na raiz do diretório temos
1. Um arquivo `c_grader.json` que é um arquivo de configuração do API Client [Insomnium](https://github.com/ArchGPT/insomnium) (que é um fork do Insomnia, mas sem a parte online) que mostra todas as rotas com requisições já prontas para elas. A documentação da api também é feita aqui, e você pode ver como cada rota funciona individualmente abrindo-as na aplicação e olhando a aba `docs`.
2. O DER (Diagrama de entidades e relações) da nossa database, que demonstra quais as tabelas que existem e a relação entre elas. Foi feita no site [Draw.io](https://app.diagrams.net/) e recomendo abrir a imagem dentro do arquivo para facilitar a leitura, visto que o png tem alguns defeitos de visualização.
