# Spielplatz

Ein online-Spielplatz für DDP.

Inspiriert von anderen Sprachen wie [Go](https://go.dev/play/), [Rust](https://play.rust-lang.org/?version=stable&mode=debug&edition=2021) und [Dart](https://dartpad.dev/?).

![showcase](img/showcase.png)

## Starten
### Vorraussetzungen
* Go version 1.20.0
* Npm

### Ausführen
Um das Programm zu starten führt man `run.sh` aus.
Das Makefile sollte dann alle Abhängigkeiten automatisch installieren (eventuell muss das sudo Passwort angegeben werden).

### Konfiguration
Man kann im root des Projektes eine `config.json` Datei erstellen um das Programm einszustellen.
Die standart Konfigurationsdatei sieht so aus.
```json
{
	"cpu_limit_percent": 50,
	"exe_cache_duration": "1m",
	"memory_limit_bytes": 4294967296,
	"port": "8080",
	"run_timeout": "30s"
}
```
