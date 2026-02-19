---
id: ko-fb9e
status: closed
deps: []
links: []
created: 2026-02-19T02:22:19Z
type: task
priority: 2
---
# Add ModTime field to Ticket struct, populate from filesystem in ListTickets

## Notes

**2026-02-19 02:22:32 UTC:** Add ModTime time.Time to Ticket struct (yaml:"-", not serialized). In ListTickets (ticket.go:216), after reading the DirEntry, call e.Info() to get FileInfo and set t.ModTime = info.ModTime(). LoadTicket also needs it — os.Stat the file path after ReadFile. This is the foundation for mtime-based sorting.
