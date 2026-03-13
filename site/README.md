
## Directory structure
```
site/
├─ src/
│  ├─ lib/
│  │  ├─ assets/    - Assets and favicon
│  │  ├─ components/
│  │  │  ├─ common/    - Components which are used by other components
│  │  │  └─ core/      - Complex Components used by the playground
│  │  └─ data/      - Monaco-Editor configuration data
│  ├─ routes/     - Filesystem based routing
│  └─ app.html    - Base html file. Contains css reset and variables
└─ static/    - Other static data, like webmanifest and robots.txt
```

## Running and Building
Prerequisites: 
- Install packages `npm install`
- Generate cloudflare types: `npm run cf-typegen` (prevents error is tsconfig)
- The backend needs to be running


To debug locally: `npm run dev`\
To build the output files: `npm run build`

