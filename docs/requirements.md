# Project Requirements
## Functional Requirements
**P0**
- Fetch and display a feed's item
- Item deduplication
- Subscribe/unsubscribe capability
- Multi-feed support

**P1**
- Multi-user support
- Mark an item read/unread

**P2**
- Add an item to favorites (bookmarking)
- Add tags to an item
- OPML import

## Non-Functional Requirements
**P0**
- Basic monitoring
- Broken feeds don't affect polling healthy feeds

**P2**
- Data retained for 1 year (P2, because it's a pet project. Don't really care about the data size since it's going to be manageable)


## Out of Scope
- Mobile app
- Item sharing (at least for now)
- Account deletion