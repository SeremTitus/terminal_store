package main

import (
    "bufio"
    "bytes"
    "encoding/json"
    "fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"terminal_store/pkg/env"
	"terminal_store/pkg/models"
)

func serverUp(baseURL string) bool {
    client := http.Client{Timeout: 800 * time.Millisecond}
    resp, err := client.Get(baseURL + "/health")
    if err != nil {
        return false
    }
    resp.Body.Close()
    return resp.StatusCode == 200
}

func getJSON[T any](url string, out *T) error {
    client := http.Client{Timeout: 3 * time.Second}
    resp, err := client.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        return fmt.Errorf("server error: %s", resp.Status)
    }
    return json.NewDecoder(resp.Body).Decode(out)
}

func sendJSON[T any](method, url string, body any, out *T) error {
    b, err := json.Marshal(body)
    if err != nil {
        return err
    }
    req, err := http.NewRequest(method, url, bytes.NewReader(b))
    if err != nil {
        return err
    }
    req.Header.Set("Content-Type", "application/json")
    client := http.Client{Timeout: 4 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 300 {
        return fmt.Errorf("server error: %s", resp.Status)
    }
    return json.NewDecoder(resp.Body).Decode(out)
}

func readLine(reader *bufio.Reader, prompt string) string {
    fmt.Print(prompt)
    line, _ := reader.ReadString('\n')
    return strings.TrimSpace(line)
}

func readInt(reader *bufio.Reader, prompt string) (int64, error) {
    for {
        text := readLine(reader, prompt)
        v, err := strconv.ParseInt(text, 10, 64)
        if err == nil {
            return v, nil
        }
        fmt.Println("Please enter a valid integer.")
    }
}

func readFloat(reader *bufio.Reader, prompt string) (float64, error) {
    for {
        text := readLine(reader, prompt)
        v, err := strconv.ParseFloat(text, 64)
        if err == nil {
            return v, nil
        }
        fmt.Println("Please enter a valid number.")
    }
}

func menu(baseURL string) {
    reader := bufio.NewReader(os.Stdin)

    for {
        fmt.Println("\n--- MENU ---")
        fmt.Println("1) List products")
        fmt.Println("2) Add product")
        fmt.Println("3) Update stock")
        fmt.Println("4) List customers")
        fmt.Println("5) Add customer")
        fmt.Println("6) Create order")
        fmt.Println("7) View orders")
        fmt.Println("0) Exit")
        fmt.Print("> ")

        line, _ := reader.ReadString('\n')
        choice := strings.TrimSpace(line)

        switch choice {
        case "1":
            var products []models.Product
            if err := getJSON(baseURL+"/products", &products); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
            fmt.Fprintln(tw, "ID\tNAME\tPRICE\tSTOCK\tCREATED")
            for _, p := range products {
                fmt.Fprintf(tw, "%d\t%s\t%.2f\t%d\t%s\n", p.ID, p.Name, p.Price, p.Stock, p.CreatedAt.Format(time.RFC3339))
            }
            tw.Flush()
        case "2":
            name := readLine(reader, "Product name: ")
            price, _ := readFloat(reader, "Price: ")
            stock, _ := readInt(reader, "Stock: ")
            req := map[string]any{
                "name":  name,
                "price": price,
                "stock": stock,
            }
            var created models.Product
            if err := sendJSON(http.MethodPost, baseURL+"/products", req, &created); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            fmt.Printf("Created product #%d\n", created.ID)
        case "3":
            pid, _ := readInt(reader, "Product ID: ")
            stock, _ := readInt(reader, "New stock: ")
            req := map[string]any{
                "stock": stock,
            }
            var updated models.Product
            if err := sendJSON(http.MethodPatch, baseURL+"/products/"+strconv.FormatInt(pid, 10)+"/stock", req, &updated); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            fmt.Printf("Updated product #%d stock=%d\n", updated.ID, updated.Stock)
        case "4":
            var customers []models.Customer
            if err := getJSON(baseURL+"/customers", &customers); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            tw := tabwriter.NewWriter(os.Stdout, 0, 2, 2, ' ', 0)
            fmt.Fprintln(tw, "ID\tNAME\tPHONE\tCREATED")
            for _, c := range customers {
                fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", c.ID, c.Name, c.Phone, c.CreatedAt.Format(time.RFC3339))
            }
            tw.Flush()
        case "5":
            name := readLine(reader, "Customer name: ")
            phone := readLine(reader, "Phone (optional): ")
            req := map[string]any{
                "name":  name,
                "phone": phone,
            }
            var created models.Customer
            if err := sendJSON(http.MethodPost, baseURL+"/customers", req, &created); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            fmt.Printf("Created customer #%d\n", created.ID)
        case "6":
            cid, _ := readInt(reader, "Customer ID: ")
            items := make([]map[string]any, 0)
            for {
                pidText := readLine(reader, "Product ID (blank to finish): ")
                if pidText == "" {
                    break
                }
                pid, err := strconv.ParseInt(pidText, 10, 64)
                if err != nil {
                    fmt.Println("Invalid product id.")
                    continue
                }
                qty, _ := readInt(reader, "Qty: ")
                items = append(items, map[string]any{
                    "product_id": pid,
                    "qty":        qty,
                })
            }
            req := map[string]any{
                "customer_id": cid,
                "items":       items,
            }
            var created models.Order
            if err := sendJSON(http.MethodPost, baseURL+"/orders", req, &created); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            total := 0.0
            for _, it := range created.Items {
                total += it.PriceEach * float64(it.Qty)
            }
            fmt.Printf("Created order #%d with %d items (total $%.2f)\n", created.ID, len(created.Items), total)
        case "7":
            var orders []models.Order
            if err := getJSON(baseURL+"/orders", &orders); err != nil {
                fmt.Println("Error:", err)
                continue
            }
            for _, o := range orders {
                fmt.Printf("Order #%d customer=%d created=%s\n", o.ID, o.CustomerID, o.CreatedAt.Format(time.RFC3339))
                for _, it := range o.Items {
                    fmt.Printf("  item #%d product=%d qty=%d price=%.2f\n", it.ID, it.ProductID, it.Qty, it.PriceEach)
                }
            }
        case "0":
            return
        default:
            fmt.Println("Invalid choice")
        }
    }
}

func main() {
	_ = env.Load(".env")
	baseURL := os.Getenv("SERVER_URL")
    if baseURL == "" {
        baseURL = "http://localhost:8080"
    }

    if !serverUp(baseURL) {
        fmt.Println("Server still not reachable.")
        return
    }

    menu(baseURL)
}
