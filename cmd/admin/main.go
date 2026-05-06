package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"alteon-api/internal/config"
	"alteon-api/internal/storage"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cfg := config.Load()

	s, err := storage.Open(cfg.DB.URL)
	if err != nil {
		exit(err)
	}
	defer s.Close()

	alteons := storage.NewAlteonsRepo(s)
	tokens := storage.NewTokensRepo(s)
	ctx := context.Background()

	switch os.Args[1] {
	case "add-alteon":
		if len(os.Args) != 6 {
			fmt.Fprintln(os.Stderr, "uso: admin add-alteon <name> <url> <user> <pass>")
			os.Exit(2)
		}
		id, err := alteons.Create(ctx, storage.Alteon{
			Name:               os.Args[2],
			BaseURL:            os.Args[3],
			Username:           os.Args[4],
			Password:           os.Args[5],
			InsecureSkipVerify: true,
			Enabled:            true,
		})
		if err != nil {
			exit(err)
		}
		fmt.Printf("alteon creado: id=%d name=%s\n", id, os.Args[2])

	case "list-alteons":
		list, err := alteons.List(ctx)
		if err != nil {
			exit(err)
		}
		fmt.Printf("%-4s %-20s %-35s %-10s %s\n", "ID", "NAME", "URL", "USER", "ENABLED")
		for _, a := range list {
			fmt.Printf("%-4d %-20s %-35s %-10s %v\n", a.ID, a.Name, a.BaseURL, a.Username, a.Enabled)
		}

	case "remove-alteon":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "uso: admin remove-alteon <name>")
			os.Exit(2)
		}
		if err := alteons.DeleteByName(ctx, os.Args[2]); err != nil {
			exit(err)
		}
		fmt.Println("alteon removido")

	case "enable-alteon", "disable-alteon":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "uso: admin "+os.Args[1]+" <name>")
			os.Exit(2)
		}
		enabled := os.Args[1] == "enable-alteon"
		if err := alteons.SetEnabled(ctx, os.Args[2], enabled); err != nil {
			exit(err)
		}
		fmt.Println("ok")

	case "create-token":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "uso: admin create-token <name>")
			os.Exit(2)
		}
		plain, id, err := tokens.Create(ctx, os.Args[2])
		if err != nil {
			exit(err)
		}
		fmt.Printf("token creado: id=%d name=%s\n", id, os.Args[2])
		fmt.Println("GUARDA ESTE TOKEN (no se vuelve a mostrar):")
		fmt.Println(plain)

	case "list-tokens":
		list, err := tokens.List(ctx)
		if err != nil {
			exit(err)
		}
		fmt.Printf("%-4s %-20s %-25s %-25s %s\n", "ID", "NAME", "CREATED", "LAST USED", "REVOKED")
		for _, t := range list {
			last := "-"
			if t.LastUsedAt != nil {
				last = t.LastUsedAt.Format(time.RFC3339)
			}
			fmt.Printf("%-4d %-20s %-25s %-25s %v\n", t.ID, t.Name, t.CreatedAt.Format(time.RFC3339), last, t.Revoked)
		}

	case "revoke-token":
		if len(os.Args) != 3 {
			fmt.Fprintln(os.Stderr, "uso: admin revoke-token <id>")
			os.Exit(2)
		}
		id, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil {
			exit(err)
		}
		if err := tokens.Revoke(ctx, id); err != nil {
			exit(err)
		}
		fmt.Println("token revocado")

	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `alteon-api admin

Comandos:
  add-alteon <name> <url> <user> <pass>
  list-alteons
  remove-alteon <name>
  enable-alteon <name>
  disable-alteon <name>
  create-token <name>
  list-tokens
  revoke-token <id>

Env:
  DATABASE_URL  DSN de postgres (default: postgres://alteon:alteon@localhost:5432/alteon?sslmode=disable)`)
}

func exit(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}
