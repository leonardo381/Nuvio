package main

import (
    "database/sql"
    "fmt"
    _ "modernc.org/sqlite"
)

func main() {
    db, err := sql.Open("sqlite", `C:\Users\Leo\Documents\Nuvio\NuvioCMS\Nuvio\pb_data\data.db`)
    if err != nil { panic(err) }
    defer db.Close()

    rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type='table' ORDER BY name`)
    if err != nil { panic(err) }
    defer rows.Close()
    fmt.Println("TABLES:")
    for rows.Next() {
        var name string
        if err := rows.Scan(&name); err != nil { panic(err) }
        if name == "pbc_1661203500" || name == "pbc_1661203400" || name == "pbc_1661203600" || name == "_collections" {
            fmt.Println(" -", name)
        }
    }

    fmt.Println("\nCAMPAIGNS columns:")
    crows, err := db.Query(`PRAGMA table_info(pbc_1661203500)`)
    if err != nil { panic(err) }
    defer crows.Close()
    for crows.Next() {
        var cid int
        var name, ctype string
        var notnull int
        var dflt sql.NullString
        var pk int
        if err := crows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil { panic(err) }
        fmt.Printf(" - %d %s %s\n", cid, name, ctype)
    }

    fmt.Println("\nLATEST campaigns:")
    q := `SELECT id, status, subject, website, recipientsType, recipientsIds, recipientsCount, sentAt, created FROM pbc_1661203500 ORDER BY created DESC LIMIT 10`
    rrows, err := db.Query(q)
    if err != nil { panic(err) }
    defer rrows.Close()
    for rrows.Next() {
        var id, status, subject, website, recipientsType, recipientsIds, sentAt, created sql.NullString
        var recipientsCount sql.NullInt64
        if err := rrows.Scan(&id, &status, &subject, &website, &recipientsType, &recipientsIds, &recipientsCount, &sentAt, &created); err != nil { panic(err) }
        fmt.Printf("id=%s status=%s rType=%s rIds=%q rCount=%d subjectEmpty=%v website=%s created=%s\n", id.String, status.String, recipientsType.String, recipientsIds.String, recipientsCount.Int64, subject.String=="", website.String, created.String)
    }
}
