### About
Minimilistic Resource Pool manager with following feature
- Allows to define pool size (0 to 255)
- Simple 2 APIs to interact with pool
  - `Get` - to get resource from the pool, it also internally handles creation if resource
  - `Release` - to put resource back to the pool, it also internally handles reset resource (if any)
