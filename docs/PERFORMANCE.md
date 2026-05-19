## 2026-05-19

| **Legacy Architecture (JIT Processing)** | **New Architecture (Read-Through Cache)** |
| :--- | :--- |
| Processed pages synchronously on HTTP request | Background worker pre-processes pages to disk |
| Unbounded concurrency saturated CPU | Throttled worker yields CPU (500ms sleep) |
| HTTP router starved under load | Router serves static files instantly |
| **Peak CPU: 327% (99s active time)** | **Peak CPU: 1.23% (370ms active time)** |

### Before (JIT)
```text
Duration: 30.28s, Total samples = 99.27s (327.85%)
      flat  flat%   sum%        cum   cum%
    12.49s 12.58% 12.58%     21.07s 21.22%  golang.org/x/image/draw.nnInterpolator.scale_RGBA_RGBA64Image_Src
    10.94s 11.02% 23.60%     37.95s 38.23%  image/png.(*decoder).readImagePass
     9.06s  9.13% 32.73%     25.15s 25.33%  image/jpeg.(*encoder).writeBlock
```

### After (Async Worker + SQLite Queue)
```text
Duration: 30s, Total samples = 370ms ( 1.23%)
      flat  flat%   sum%        cum   cum%
     170ms 45.95% 45.95%      170ms 45.95%  internal/runtime/syscall/linux.Syscall6
      40ms 10.81%  ...         ...          modernc.org/sqlite/lib...
      30ms  8.11%  ...         ...          modernc.org/memory.(*Allocator)...
```
