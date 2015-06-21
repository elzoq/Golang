package main

import "fmt"

// PersonInfo ��һ��������ϸ������Ϣ������
type PersonInfo struct {
	ID string
	Name string
	Address string
}

func main() {
	var personDB map[string] PersonInfo
	personDB = make(map[string] PersonInfo)
	
	// ����� map ����뼸������
	personDB["12345"] = PersonInfo{"12345", "Tom", "Room 203,..."}
	personDB["1"] = PersonInfo{"1", "Jack", "Room 101,..."}
	
	// ����� map ���� key Ϊ"1234"����Ϣ
	person, ok := personDB["1234"]
	
	// ok ��һ�����ص� bool ��,���� true ��ʾ�ҵ��˶�Ӧ������
	if ok {
		fmt.Println("Found person", person.Name, "with ID 1234.")
	} else {
		fmt.Println("Did not find person with ID 1234.")
	}
}