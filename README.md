# Spielplatz

Ein online-Spielplatz um [DDP](https://github.com/DDP-Projekt/Kompilierer) auszuprobieren.\
Erreichbar unter https://spiel.ddp.im

Inspiriert von anderen Sprachen wie [Go](https://go.dev/play/), [Rust](https://play.rust-lang.org/?version=stable&mode=debug&edition=2021) und [Dart](https://dartpad.dev/?).

## Lokal Ausführen
### Vorraussetzungen
* [Go](https://go.dev/doc/install) (mindestens version 1.26.1)
* npm

### Installieren
1. Die Spielplatz und [Kompilierer](https://github.com/DDP-Projekt/Kompilierer) Repositories klonen
2. Den Kompilierer bauen
3. (auf Linux) `make install-dependencies` ausführen. (auf Windows) `make unsec_main.o`
4. Frontend-Abhängigkeiten installieren: `npm --prefix site install` und `npm --prefix site run cf-typegen`
5. `site/.env` Datei mit der Variable `PUBLIC_BACKEND_HOST` erstellen und auf Host+Port (oder Domain) des Backends setzen (z. B. `PUBLIC_BACKEND_HOST = localhost:8080`)

### Ausführen
Backend und Frontend laufen lokal getrennt:

1. Backend starten:
`go run -ldflags "-X main.DDPVERSION=dev" ./server/`

2. Frontend starten:
`npm --prefix site run dev`

Das Frontend läuft standardmäßig auf `http://localhost:5173` und leitet `/api` an das Backend (`PUBLIC_BACKEND_HOST`) weiter.

## Backend via Docker Ausführen
### Vorraussetzungen

* [Docker](https://docs.docker.com/get-docker/)

### Installieren

1. Das Spielplatz Repository klonen
2. Die entsprechende LLVM Version von https://github.com/llvm/llvm-project/releases/tag/llvmorg-12.0.0 herunterladen 
3. Docker starten
4. `./docker-build.sh <llvm_archive>` auführen, wobei `<llvm_archive>` der Pfad zur heruntergeladenen LLVM Version ist

Bsp.:
```bash
wget https://github.com/llvm/llvm-project/releases/download/llvmorg-12.0.0/clang+llvm-12.0.0-x86_64-linux-gnu-ubuntu-20.04.tar.xz
./docker-build.sh clang+llvm-12.0.0-x86_64-linux-gnu-ubuntu-20.04.tar.xz
```

### Ausführen

Das Docker Image kann mit `docker run -p 8080:8080 ddp-spielplatz` als container gestartet werden.
Der Spielplatz Backend sollte jetzt unter http://localhost:8080/ erreicht werden können.

## Konfiguration
Der Backend lässt sich über eine `config.json` Datei einstellen.
Die standart Konfigurationsdatei sieht so aus.
```json
{
	"certpath": "",
	"cpu_limit_percent": 50,
	"exe_cache_duration": 60000000000,
	"keypath": "",
	"log_level": "INFO",
	"max_concurrent_processes": 50,
	"max_source_code_log_length": 100,
	"memory_limit_bytes": 4294967296,
	"port": "8080",
	"pprof": false,
	"process_aquire_timeout": 3000000000,
	"run_timeout": 60000000000,
	"share_db_path": "./share_links.db",
	"usehttps": false
}
```
