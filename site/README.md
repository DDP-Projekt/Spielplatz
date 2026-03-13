## Verzeichnisstruktur
```
site/
├─ src/
│  ├─ lib/
│  │  ├─ assets/    - Assets und Favicon
│  │  ├─ components/
│  │  │  ├─ common/    - Komponenten, die von anderen Komponenten verwendet werden
│  │  │  └─ core/      - Komplexe Komponenten für den Playground
│  │  └─ data/      - Konfigurationsdaten für den Monaco-Editor
│  ├─ routes/     - Dateisystembasiertes Routing
│  └─ app.html    - Basis-HTML-Datei. Enthält CSS-Reset und Variablen
└─ static/    - Weitere statische Daten, z. B. Webmanifest und robots.txt
```

## Ausführen und Bauen
Voraussetzungen: 
- Pakete installieren: `npm install`
- Cloudflare-Typen generieren: `npm run cf-typegen` (verhindert Fehler in der tsconfig)
- Eine `.env` Datei mit der Variable `PUBLIC_BACKEND_HOST` erstellen und auf Host+Port (oder Domain) des Backends setzen (z. B. `PUBLIC_BACKEND_HOST = localhost:8080`)
- Das Backend muss laufen


Für lokales Debugging: `npm run dev`\
Zum Bauen der Ausgabedateien: `npm run build`

