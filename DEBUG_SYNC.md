# 🐛 Debug Guide - Watch Party Sync

## Vấn đề: Video không sync giữa các tabs

### ✅ Đã fix

**Vấn đề:** Khi một tab play/pause, tab khác không phản ứng

**Nguyên nhân:** WebSocket messages từ server không được xử lý để update `videoState`

**Giải pháp:** 
1. ✅ Update `videoState` khi nhận play/pause/seek messages
2. ✅ Thêm console.log để debug
3. ✅ Loại bỏ `isSyncing` check trong useEffect
4. ✅ Tăng tolerance cho seek từ 1s → 2s

## 🧪 Cách test sau khi fix

### Bước 1: Restart React App
```bash
# Nếu React đang chạy, nhấn Ctrl+C
cd "f:\smbweb test\movieapp\movie"
npm run dev
```

### Bước 2: Đảm bảo Go backend đang chạy
```bash
cd "f:\smbweb test\movieapp"
go run .
```

### Bước 3: Mở 2 browser tabs
1. Tab 1: `http://localhost:5173/watch-party-test`
2. Tab 2: `http://localhost:5173/watch-party-test`

### Bước 4: Create room ở Tab 1
- Nhập username (ví dụ: "User1")
- Click "Create New Room"
- Copy Room ID

### Bước 5: Join room ở Tab 2
- Click "Join Existing Room"
- Paste Room ID
- Nhập username khác (ví dụ: "User2")

### Bước 6: Test sync
- ✅ Play ở Tab 1 → Tab 2 auto play
- ✅ Pause ở Tab 2 → Tab 1 auto pause
- ✅ Seek (kéo timeline) ở Tab 1 → Tab 2 jump đến cùng vị trí

## 🔍 Debug với Browser Console

Mở DevTools (F12) → Console, bạn sẽ thấy:

**Tab 1 (khi click play):**
```
Connected to watch party
Video state changed: {isPlaying: true, currentTime: 5.2, ...}
```

**Tab 2 (khi nhận sync):**
```
User1 play {currentTime: 5.2}
Video state changed: {isPlaying: true, currentTime: 5.2, ...}
Playing video
```

## 🐛 Nếu vẫn không sync

### Check 1: WebSocket connected?
Console sẽ có:
```
Connected to watch party
```

Nếu không thấy → Check Go backend đang chạy

### Check 2: Cùng Room ID?
Console sẽ hiển thị Room ID. Verify 2 tabs cùng room.

### Check 3: Messages được nhận?
Khi Tab 1 play, Console Tab 2 phải show:
```
User1 play {currentTime: ...}
```

Nếu không → Check Go backend logs

### Check 4: Go Backend Logs
Terminal chạy Go backend sẽ show:
```
Client User1 joined room abc123
Room abc123: User1 played at 5.20
```

## 💡 Known Issues

### Issue 1: Video không load (m3u8)
- **Nguyên nhân:** CORS hoặc m3u8 server không available
- **Giải pháp:** Check browser console, có thể cần thay link video khác

### Issue 2: Video delay khi sync
- **Normal:** Có delay nhỏ (< 1s) do network latency
- **Không normal:** Delay > 3s → Check network speed

### Issue 3: Seek không chính xác
- **Normal:** Tolerance là 2 giây
- **Fix:** Giảm tolerance trong code nếu cần chính xác hơn

## 📊 Expected Flow

```
Tab 1 User clicks PLAY
    ↓
Handle Play → Send WebSocket message
    ↓
Go Server receives play message
    ↓
Server broadcasts to all clients in room
    ↓
Tab 2 receives play message
    ↓
Update videoState (isPlaying: true)
    ↓
useEffect triggers → video.play()
    ↓
✅ Video plays on Tab 2!
```
