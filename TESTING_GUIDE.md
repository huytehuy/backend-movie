# 🧪 Hướng dẫn Test Watch Party

## Bước 1: Start Go Backend Server

Mở terminal và chạy:

```bash
cd "f:\smbweb test\movieapp"
go run .
```

Bạn sẽ thấy:
```
Server starting on port 8080
Video streaming: http://localhost:8080/api/videos/
WebSocket: ws://localhost:8080/api/rooms/{roomId}/ws
```

## Bước 2: Thêm Video Test

Copy một file video vào thư mục `videos/`:

```bash
# Ví dụ:
copy "D:\Downloads\sample.mp4" "f:\smbweb test\movieapp\videos\sample1.mp4"
```

## Bước 3: Start React Dev Server

Mở terminal mới và chạy:

```bash
cd "f:\smbweb test\movieapp\movie"
npm run dev
```

React app sẽ chạy tại: `http://localhost:5173`

## Bước 4: Truy cập Test Page

Mở browser và truy cập:
```
http://localhost:5173/watch-party-test
```

## Bước 5: Test Watch Party

### Test 1: Tạo Room
1. Nhập username của bạn
2. Click "Create New Room"
3. Room sẽ được tạo và hiển thị Room ID

### Test 2: Join Room từ Tab khác
1. Copy Room ID từ tab đầu tiên
2. Mở tab/window mới: `http://localhost:5173/watch-party-test`
3. Click "Join Existing Room"
4. Paste Room ID

### Test 3: Test Video Sync
Với 2 tabs đang mở (cùng room):
- ✅ **Play**: Click play ở tab 1 → video sẽ play ở tab 2
- ✅ **Pause**: Click pause ở tab 1 → video sẽ pause ở tab 2
- ✅ **Seek**: Kéo timeline ở tab 1 → video sẽ jump đến cùng vị trí ở tab 2
- ✅ **User List**: Sẽ thấy danh sách users trong phòng

## 🔍 Debug Tips

### Check Go Backend Logs
Terminal chạy Go server sẽ hiển thị:
```
Room created: abc123 for movie 1 by User1
Client User1 joined room abc123
Room abc123: User1 played at 10.50
```

### Check Browser Console
Mở DevTools (F12) → Console, sẽ thấy:
```
Connected to watch party
User2 play {currentTime: 10.5}
```

### Check Network Tab
DevTools → Network → WS (WebSocket):
- Sẽ thấy WebSocket connection
- Click vào để xem messages trao đổi

## 🐛 Troubleshooting

### Lỗi: "Failed to create room"
- ✅ Check Go backend đang chạy (`http://localhost:8080/api/health`)
- ✅ Check CORS settings trong `main.go`

### Video không sync
- ✅ Check browser console có lỗi WebSocket không
- ✅ Verify cả 2 tabs cùng roomId
- ✅ Check Go backend logs xem có nhận message không

### Video không load
- ✅ Check file video tồn tại trong `videos/sample1.mp4`
- ✅ Thử truy cập trực tiếp: `http://localhost:8080/api/videos/sample1.mp4`

## 📊 Expected Behavior

**Khi User A click play:**
1. Browser A → WebSocket message → Go server
2. Go server → Broadcast → All connected browsers
3. Browser B nhận message → Auto play video

**Timeline:**
```
Browser A:  [Play Click] ──▶ WebSocket ──▶ Server
                                            │
Server:     Broadcast play message         │
                                            │
Browser B:  ◀──── WebSocket ◀──────────────┘
            [Auto Play]
```

## 🎯 Test Checklist

- [ ] Go server starts successfully
- [ ] React dev server starts successfully
- [ ] Can access test page
- [ ] Can create room
- [ ] Can join room from another tab
- [ ] Play sync works
- [ ] Pause sync works
- [ ] Seek sync works
- [ ] User list updates
- [ ] WebSocket reconnects after disconnect

## ✨ Next Steps

Sau khi test thành công, bạn có thể:
1. Tích hợp watch party vào trang xem phim chính
2. Thêm chat feature
3. Thêm room password
4. Thêm host controls (chỉ host mới control video)
5. Deploy lên production
