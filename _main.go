// 아예 파일 있고 없고가 다르다면 각 폴더별 각각에 있는 파일만 표시
// 같은 이름, md5가 다르다면 공통 영역에 표시

var md5s map[string][][16]byte // 이거보다 sum 이 더 나음

func diffMd5s(check, contrast [][16]byte) [][16]byte {
	diff := make([][16]byte, 0)
	for _, e := range check {
		if !containsMd5(contrast, e) {
			diff = append(diff, e)
		}
	}
	return diff
}

func containsMd5(s [][16]byte, e [16]byte) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
var diff1, diff2 [][16]byte
func (){
	if existPath, ok := marked[id]; ok {
		diff1 = diffMd5s(md5s[existPath], md5s[songPath])
		diff2 = diffMd5s(md5s[songPath], md5s[existPath])
		if len(diff1) == 0 && len(diff2) == 0 {
			// 같은 거는 이전 거를 삭제
		} else {
			// 다른 거는 이전 거를 이동
		}
	} else {
		marked[id] = songPath
	}
}
