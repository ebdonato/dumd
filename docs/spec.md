# Especificação Técnica: DuMD - Visualizador de Markdown Ultra Minimalista

Esta especificação descreve os requisitos, a arquitetura e os detalhes de implementação para o **DuMD**, um visualizador de Markdown desktop ultra-minimalista, portátil e controlado principalmente por teclado, focado em alta legibilidade e leveza.

## Visão Geral do Produto

O **DuMD** é uma aplicação desktop cujo único objetivo é renderizar arquivos Markdown locais de forma elegante e limpa. Ele foi projetado para usuários que valorizam o minimalismo, a eficiência do teclado (atalhos Vim/leitura rápida) e a facilidade de distribuição (binário único, sem instaladores complexos).

### Diferenciais Principais

- **Executável Único:** Um único arquivo compilado contendo o motor em Go e a interface Webview embutida.
- **Foco no Teclado:** Controle de navegação inspirado no Vim e rolagem fluida por espaço.
- **Minimalismo Visual:** Interface despida de menus complexos, utilizando apenas a moldura de janela nativa e um botão flutuante discreto para personalização básica.

## Requisitos Funcionais (RF)

| ID       | Nome do Requisito               | Descrição                                                                                                                                                                |
| :------- | :------------------------------ | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **RF01** | **Carregamento de Arquivos**    | O aplicativo deve abrir e ler um arquivo Markdown (`.md`) passado como argumento por linha de comando (ex: `dumd readme.md`).                                          |
| **RF02** | **Renderização de Markdown**    | O conteúdo do arquivo deve ser convertido em HTML5 semântico e estilizado com uma tipografia altamente legível.                                                          |
| **RF03** | **Controles de Janela Nativos** | A janela deve reter as decorações padrão do sistema operacional hospedeiro (botões de fechar, minimizar e maximizar).                                                    |
| **RF04** | **Menu de Configuração Rápida** | Um painel flutuante ou botão discreto (ex: canto inferior direito) deve permitir a troca de zoom e alteração do tema visual.                                             |
| **RF05** | **Suporte a Temas**             | O sistema deve suportar 3 temas de leitura: **Claro** (papel limpo), **Escuro** (alto contraste e amigável aos olhos à noite) e **Sépia** (conforto térmico de leitura). |
| **RF06** | **Controle de Zoom**            | O usuário deve poder ajustar o tamanho da fonte (ex: 80% a 150%) através do menu ou de atalhos.                                                                          |
| **RF07** | **Navegação por Teclado**       | Implementar mapeamento completo de teclas para rolagem e controle do app.                                                                                                |

## Requisitos Não-Funcionais (RNF)

| ID        | Requisito Não-Funcional                | Meta de Desempenho / Restrição                                                                                                        |
| :-------- | :------------------------------------- | :------------------------------------------------------------------------------------------------------------------------------------ |
| **RNF01** | **Portabilidade**                      | O aplicativo deve compilar nativamente para Windows (x64), macOS (Apple Silicon/Intel) e Linux (x64).                                 |
| **RNF02** | **Tamanho do Executável**              | O tamanho final do binário compilado e otimizado não deve exceder **18 MB**.                                                          |
| **RNF03** | **Consumo de Memória**                 | O consumo de memória RAM em repouso deve ser inferior a **60 MB**.                                                                    |
| **RNF04** | **Distribuição de Arquivo Único**      | Todo o frontend (HTML, CSS, JS) e recursos adicionais devem ser embutidos no binário Go usando a biblioteca nativa `embed` (Go 1.16+).|
| **RNF05** | **Sem Dependência de Runtime Externo** | O usuário final não deve precisar instalar Node.js, Python ou bibliotecas extras no sistema para rodar o app.                         |

## Arquitetura Tecnológica

A aplicação utilizará um modelo híbrido de **Backend em Go** e **Frontend em HTML/CSS/JS** encapsulado por uma WebView nativa do sistema operacional.

```text
+-------------------------------------------------------------+
|                       MOLDURA DA JANELA                     |
|  +-------------------------------------------------------+  |
|  | [Wails Webview] HTML / CSS / JS (Vanilla)             |  |
|  |                                                       |  |
|  | - Captura de eventos de teclado (Vim / Espaço / ESC)  |  |
|  | - Renderização do HTML estilizado                     |  |
|  | - Gerenciamento de Tema (Claro, Escuro, Sépia)        |  |
|  | - Ajuste de Zoom Dinâmico                             |  |
|  +-------------------------------------------------------+  |
|                              | (Comunicação IPC Bidirecional)
|                              v                              |
|  +-------------------------------------------------------+  |
|  | [Backend em Go]                                       |  |
|  |                                                       |  |
|  | - Leitura do arquivo .md (I/O)                        |  |
|  | - Parser de Markdown -> HTML (Goldmark)               |  |
|  | - Chamadas do Sistema (Fechar Janela ao receber ESC)  |  |
|  +-------------------------------------------------------+  |
+-------------------------------------------------------------+
```

### Componentes de Software

1. **Framework Principal (Wails v2):** Conecta o backend Go ao frontend de forma nativa e otimizada utilizando os motores Web nativos:
   - Windows: _Microsoft Edge WebView2 (Chromium)_
   - macOS: _WebKit (WKWebView)_
   - Linux: _WebKitGTK_
2. **Parser de Markdown (Goldmark):** Parser extremamente rápido escrito em Go puro que segue a especificação CommonMark.
3. **Frontend (HTML5/CSS3/Vanilla JS):** O frontend não usará React, Vue ou Angular para manter o peso mínimo de carregamento. O CSS definirá o design responsivo, estilos de tipografia fluida e variáveis de tema.

## Mapeamento de Atalhos e Teclado

O controle por teclado é prioritário para a usabilidade do aplicativo.

| Tecla / Combinação | Ação Executada                                    | Mecanismo de Implementação                                                        |
| :----------------- | :------------------------------------------------ | :-------------------------------------------------------------------------------- |
| `Escape` (ESC)     | Fecha o aplicativo imediatamente                  | Captura no JS -> Chama `runtime.WindowClose(ctx)` no Go.                          |
| `Espaço` (Space)   | Rola a tela para baixo (80% da altura visível)    | Evento JS `window.scrollBy({top: window.innerHeight * 0.8, behavior: 'smooth'})`  |
| `Shift + Espaço`   | Rola a tela para cima (80% da altura visível)     | Evento JS `window.scrollBy({top: -window.innerHeight * 0.8, behavior: 'smooth'})` |
| `j`                | Rola para baixo levemente (linha por linha)       | Evento JS `window.scrollBy({top: 50, behavior: 'auto'})`                          |
| `k`                | Rola para cima levemente (linha por linha)        | Evento JS `window.scrollBy({top: -50, behavior: 'auto'})`                         |
| `h`                | Rola para a esquerda (útil para blocos de código) | Evento JS `window.scrollBy({left: -50, behavior: 'auto'})`                        |
| `l`                | Rola para a direita (útil para blocos de código)  | Evento JS `window.scrollBy({left: 50, behavior: 'auto'})`                         |
| `Ctrl +` / `Cmd +` | Aumenta o Zoom (tamanho geral da fonte)           | Evento JS alterando a propriedade `document.body.style.fontSize`                  |
| `Ctrl -` / `Cmd -` | Diminui o Zoom (tamanho geral da fonte)           | Evento JS alterando a propriedade `document.body.style.fontSize`                  |
| `t`                | Alterna ciclicamente entre os Temas               | JS rotaciona a classe do body: `light` -> `dark` -> `sepia` -> `light`.           |

## Design de Interface (UI) e Temas

O design deve seguir o conceito de _"distração zero"_.

- **Sem margens extras:** O conteúdo do Markdown ocupará o centro da janela com uma largura máxima legível de `750px` (ou `70ch`), independente do tamanho da janela.
- **Tipografia:** Uso de fontes do sistema para evitar requisições web ou peso de arquivos de fontes embutidas (ex: Inter, Segoe UI, San Francisco ou Roboto).
- **Botão Flutuante (Menu):** Um pequeno ícone de engrenagem no canto inferior direito que aparece somente quando o mouse é movido para perto. Ao clicar, revela um menu suspenso simples (Pop-over) com os seletores de tema e nível de zoom.

### Variáveis CSS dos Temas (Tokens de Design)

```css
/* Tema Claro (Padrão) */
:root {
  --bg-color: #ffffff;
  --text-color: #24292f;
  --accent-color: #0969da;
  --code-bg: #f6f8fa;
  --border-color: #d0d7de;
}
```

```css
/* Tema Escuro */
body.theme-dark {
  --bg-color: #0d1117;
  --text-color: #c9d1d9;
  --accent-color: #58a6ff;
  --code-bg: #161b22;
  --border-color: #30363d;
}
```

```css
/* Tema Sépia */
body.theme-sepia {
  --bg-color: #f4ecd8;
  --text-color: #433422;
  --accent-color: #925d23;
  --code-bg: #e9dec4;
  --border-color: #d3c2a0;
}
```

## **7\. Estrutura do Projeto (Árvore de Diretórios)**

A árvore a seguir segue o padrão de estrutura exigido pelo framework **Wails**:

```text
dumd/
├── main.go             # Ponto de entrada do Go e setup do Wails
├── app.go              # Lógica de negócios (Bindings do Go: ler MD, fechar app)
├── wails.json          # Configurações de compilação e janela do Wails
├── go.mod              # Módulos dependentes (Wails, Goldmark)
├── go.sum
└── frontend/           # Recursos do Frontend (Embutidos via go:embed)
    ├── index.html      # Estrutura base da interface
    ├── src/
    │   ├── main.js     # Lógica de navegação de teclado, temas e chamadas Go
    │   └── style.css   # Estilos do Markdown (similares ao GitHub Readme) e temas
    └── assets/
        └── icon.png    # Ícone oficial do executável
```

## Estratégia de Implementação e Compilação

Para compilar e gerar o arquivo de distribuição único, o fluxo de desenvolvimento deve seguir os passos abaixo:

### Passo 1: Inicialização do Projeto

```shell
# Instalar a CLI do Wails
go install github.com/wailsapp/wails/v2/cmd/wails@latest

# Criar o projeto usando o template Vanilla JS
wails init -n dumd -t vanilla
```

### Passo 2: Implementação do Parser no Backend (app.go)

1. Importar o pacote `github.com/yuin/goldmark`.
2. Adicionar suporte para ler os argumentos de sistema (`os.Args`) na inicialização do ciclo de vida do app para detectar o arquivo `.md` a ser aberto.
3. Expor uma função do Go para o JS chamada `ObterMarkdownRenderizado()` que retorna a string em HTML gerada pelo Goldmark.

### Passo 3: Escuta de Eventos de Janela no JS

1. No arquivo `frontend/src/main.js`, disparar a leitura do Markdown chamando o backend Go.
2. Atualizar o `document.getElementById('content').innerHTML` com a resposta.
3. Configurar os listeners globais de teclado usando `window.addEventListener('keydown')` para capturar a navegação por teclado e chamar as rotinas nativas expostas (como fechar a janela no ESC).

### Passo 4: Compilação de Produção

Para compilar tudo e gerar o executável otimizado final com o tamanho reduzido, execute o comando correspondente ao seu sistema operacional:

```shell
# Compilar para o sistema atual com otimizações de produção
wails build -clean -ldflags "-s \-w"
```

_O parâmetro `-ldflags "-s \-w"` remove informações de debug do binário Go, diminuindo o executável final em aproximadamente 30% a 40% do seu tamanho original._
