package slices

import "sort"

// RemoveStringDuplicateUseMap 切片元素去重 list 待去重的切片
func RemoveStringDuplicateUseMap(list []string) []string {
	var data []string
	rd := map[string]struct{}{}
	for _, v := range list {
		if _, ok := rd[v]; !ok { //通过map内是否存在对应key值去添加对应切片内元素
			rd[v] = struct{}{}
			data = append(data, v)
		}
	}
	return data
}

// RemoveStringDuplicateUseCopy 切片元素去重并排序 list 待去重的切片
func RemoveStringDuplicateUseCopy(list []string) []string {
	if list == nil {
		return nil
	}
	out := make([]string, len(list))
	copy(out, list)
	sort.Strings(out)
	uniq := out[:0]
	for _, x := range out {
		if len(uniq) == 0 || uniq[len(uniq)-1] != x {
			uniq = append(uniq, x)
		}
	}
	return uniq
}
