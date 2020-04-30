package main

import (
	"bufio"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"time"
)

type Auto struct{
	X uint8
	Y uint8
	Farba string
	Smer bool
	Dlzka uint8
}

type Stav struct {
	Hlbka int
	Auta []Auto `json:"cars"`
	Kroky []int
	Mapa[6][6] uint8
}


// Vypise celu map do prikazoveho riadku.
func vypisMapu(mapa [6][6]uint8) {
	// Preiteruje celym 2d polom a vypise co sa nachadza na danom mieste
	for i := 0; i < 6; i++ {
		for j := 0; j < 6; j++ {
			fmt.Print(mapa[i][j], " ")
		}
		fmt.Print("\n")
	}
	fmt.Print("\n")
}

//vytvori prazdnu mapu velkosti 6x6
func vytvorPrazdnuMapu() [6][6]uint8 {
	mapa := [6][6]uint8{
		{0,0,0,0,0,0},
		{0,0,0,0,0,0},
		{0,0,0,0,0,0},
		{0,0,0,0,0,0},
		{0,0,0,0,0,0},
		{0,0,0,0,0,0},
	}
	return mapa
}

//Vytvori hash zo stavu - premeni obsah do retazca a pomocou vstavanej funkcie vytvori 64 bitovy hash
func hash(stav Stav) uint64{
	str := ""

	for i := 0; i < len(stav.Auta); i++ {
		str += stav.Auta[i].Farba
		str += strconv.Itoa(int(stav.Auta[i].X))
		str += strconv.Itoa(int(stav.Auta[i].Y))
		str += strconv.FormatBool(stav.Auta[i].Smer)
		str += strconv.Itoa(int(stav.Auta[i].Dlzka))
	}

	h := fnv.New64a()
	h.Write([]byte(str))
	return h.Sum64()

}

func dlsBezHash(zacStav Stav, hlbka int, finAuto string)(*Stav) {
	zacStav.Hlbka = 0
	zacStav.Mapa = vytvorMapuZAut(zacStav.Auta)
	//stavy, ktore su na okraji (este neprehladane)
	okraj := make([]Stav, 0)
	okraj = append(okraj, zacStav)
	//pocitadlo
	poc := 0

	for len(okraj) != 0 {
		//pop stavu, ktory sa ide navstivit
		prehlStav := okraj[len(okraj)-1]
		okraj = okraj[:len(okraj)-1]

		//Ak sa nasiel finalny stav, vrat ho, inak pokracuj v prehladavani
		if prehlStav.Hlbka == hlbka && is_state_final(prehlStav, finAuto) {
			fmt.Println("Searched ", poc, " nodes..")
			return &prehlStav
		} else {
			//ak je hlbka vramci ohranicenia, najdi deti a pushni do stacku
			if prehlStav.Hlbka < hlbka {
				poc += 1
				children := najdiDeti(prehlStav)
				okraj = append(okraj, children...)
			}
		}
	}

	return nil
}

//Zistuje, ci je auto mozne posunut o danu velkost kroku. Smer sa urcuje na zaklade znamienka kroku (ak je to kladne cislo,
//je to krok dopredu, inak dozadu) a orientacie auta. Ak krok nie je mozny, vrati chybu. Inak posunie auto o krok (zmeni jeho
//suradnice)
func posunAuto(auto *Auto, krok int8, stav Stav) (error){
	//Ak je auto horizontalne, kontroluje moznost pohybu len vpravo/vlavo
	if auto.Smer == true {
		//Kontroluje hranice mapy
		if int8(auto.Dlzka) + int8(auto.X) + krok > 6 || int8(auto.X) + krok < 0 {
			return errors.New("state impossible")
		} else {
			//Ak je krok kladny, kontroluje policka vpravo, inak vlavo
			if krok > 0{
				//Skontroluje, ci su volne policka od konca auta v sucasnej polohe po koniec auta v zelanej polohe
				for i := int8(auto.X) + int8(auto.Dlzka); i < int8(auto.X) + int8(auto.Dlzka) + krok; i++ {
					if stav.Mapa[auto.Y][i] != 0 {
						return errors.New("state impossible")
					}
				}
			} else {
				//kontroluje, ci su volne policka od zaciatku auta v sucasnej polohe po zaciatok auta v zelanej polohe
				for i := auto.X-1; int8(i) >= int8(auto.X) + krok; i-- {
					if stav.Mapa[auto.Y][i] != 0 {
						return errors.New("state impossible")
					}
				}
			}

			auto.X += uint8(krok)
		}
	} else {
		//kontroluje hranice mapy
		if int8(auto.Dlzka) + int8(auto.Y) + krok > 6 || int8(auto.Y) + krok < 0 {
			return errors.New("state impossible")
		} else {
			//Ak je krok kladny, kontroluje policka smerom dole, inak hore
			if krok > 0{
				//Skontroluje, ci su volne policka od konca auta v sucasnej polohe po koniec auta v zelanej polohe
				for i := int8(auto.Y) + int8(auto.Dlzka); i < int8(auto.Y) + int8(auto.Dlzka) + krok; i++ {
					if stav.Mapa[i][auto.X] != 0 {
						return errors.New("state impossible")
					}
				}
			} else {
				//kontroluje, ci su volne policka od zaciatku auta v sucasnej polohe po zaciatok auta v zelanej polohe
				for i := auto.Y-1; int8(i) >= int8(auto.Y) + krok; i-- {
					if stav.Mapa[i][auto.X] != 0 {
						return errors.New("state impossible")
					}
				}
			}

			auto.Y += uint8(krok)
		}
	}

	return nil
}

//Vytvori zo zoznamu aut s hlbkou, mapou a krokmi vykonanymi od povodneho stavu spolu s novym krokom
func vytvorStavZAut(auta []Auto, hlbka int, kroky []int, indexAuta int, posun int) (stav Stav){
	stav= Stav{
		Auta:   auta,
		Hlbka: hlbka,
		Mapa: vytvorMapuZAut(auta),
	}
	cpy := make([]int, len(kroky))
	copy(cpy, kroky)
	//vytvori sa novy krok, ktory sa prida k uz existujucemu zoznamu krokov.
	// 1 - hore; 2 - vpravo; 3 - dole; 4 - vlavo
	var krok []int
	if auta[indexAuta].Smer {

		if posun > 0 {
			krok = append(krok, 2)
		} else {
			krok = append(krok, 4)
			posun *= (-1)
		}

	} else {
		if posun > 0 {
			krok = append(krok, 3)
		} else {
			krok = append(krok, 1)
			posun *= (-1)
		}
	}
	krok = append(krok, indexAuta)
	krok = append(krok, posun)

	stav.Kroky = append(cpy, krok...)

	return stav
}

//Najde vsetkych potomkov daneho stavu
func najdiDeti(stav Stav)([]Stav){
	cars := stav.Auta
	children := []Stav{}
	//Posuva vsetky auta do vsetkych stran (pokial to je mozne)
	for i := 0; i < len(cars); i++ {
		for j := 1; j <= 4; j++ {

			cpy := make([]Auto, len(cars))
			copy(cpy, cars)
			err:= posunAuto(&cpy[i], int8(j), stav)
			//Ak err=nil, znamena to, ze je mozny posun a teda je vytvoreny novy stav
			if err == nil {
				new_node := vytvorStavZAut(cpy, stav.Hlbka+1, stav.Kroky, i, j)
				children = append(children, new_node)
			} else {
				break
			}
		}
	}
	for i := 0; i < len(cars); i++ {
		for j := 1; j <= 4; j++ {
			cpy := make([]Auto, len(cars))
			copy(cpy, cars)
			err := posunAuto(&cpy[i], (-1)*int8(j), stav)
			//Ak err=nil, znamena to, ze je mozny posun a teda je vytvoreny novy stav
			if err == nil {
				new_node := vytvorStavZAut(cpy, stav.Hlbka+1, stav.Kroky, i, j*(-1))
				children = append(children, new_node)
			} else {
				break
			}
		}
	}
	return children
}

//Vytvori zo zoznamu aut mapu rozmeru 6x6 so zapísanými indexmi áut na miestach, kde stoja
func vytvorMapuZAut(auta []Auto) [6][6]uint8 {
	mapa := vytvorPrazdnuMapu()

	for i := 0; i < len(auta); i++{
		leng := int(auta[i].Dlzka)
		for j := 0; j < leng; j++ {
			if auta[i].Smer {
				mapa[auta[i].Y][auta[i].X+uint8(j)] = uint8(i+1)
			} else {
				mapa[auta[i].Y+uint8(j)][auta[i].X] = uint8(i+1)
			}
		}
	}

	return mapa
}

//Vracia true ak je auto s farbou rovnakou ako je farba v argumente na konci mapy zprava
func is_state_final(state Stav, finAuto string) bool{
	for i := 0; i < len(state.Auta); i++ {
		if state.Auta[i].Farba == finAuto{
			if state.Auta[i].X + state.Auta[i].Dlzka == 6 {
				return true
			} else {
				return false
			}
		}
	}
	return false
}


//Prehladava stavovy priestor do hlbky s tym, ze neprehladava dalej ako po stanovenu hlbku
func dls(zacStav Stav, hlbka int, finAuto string, poc *int)(*Stav) {
	zacStav.Hlbka = 0
	zacStav.Mapa = vytvorMapuZAut(zacStav.Auta)
	//hash mapa, kde sa uchovavaju navstivene stavy
	hshMapa := make(map[uint64]int)
	//stavy, ktore su na okraji (este neprehladane)
	okraj := make([]Stav, 0)
	okraj = append(okraj, zacStav)

	//prehladavany stav
	var prehlStav Stav

	for len(okraj) != 0 {
		//pop stavu, ktory sa ide navstivit
		prehlStav = okraj[len(okraj)-1]
		okraj = okraj[:len(okraj)-1]
		h := hash(prehlStav)

		//ak sa nasiel finalny stav, vrat ho, inak pokracuj v prehladavani
		if prehlStav.Hlbka == hlbka && is_state_final(prehlStav, finAuto) {
			return &prehlStav
		} else if hshMapa[h] == 0 || hshMapa[h] > prehlStav.Hlbka{
			//ak stav este nebol navstiveny, alebo bol navstiveny, ale hlbsie -> navstiv ho
			hshMapa[hash(prehlStav)] = prehlStav.Hlbka
			//ak je hlbka vramci ohranicenia, najdi deti a pushni do stacku
			if prehlStav.Hlbka < hlbka{
				*poc += 1
				children := najdiDeti(prehlStav)
				okraj = append(okraj, children...)
			}
		}
	}

	return nil
}


//Spusta prehladavanie do hlbky s postupne sa zvysujucim ohranicenim
func najdiRiesenie(pociatocnyStav Stav, finalneAuto string) *Stav {
	pocitadlo := 0
	for i := 0; i < 50; i+=1 {

		najdene := dls(pociatocnyStav, i, finalneAuto, &pocitadlo)
		if najdene != nil {
			fmt.Println("Prehladalo sa ", pocitadlo, " stavov..")
			return najdene
		}
	}
	fmt.Println("Prehladalo sa ", pocitadlo, " stavov..")
	return nil
}

func loadConfig(index int) Stav{

	vsetky := map[int][][]interface{}{
		1: Config1,
		2: Config2,
		3: Config3,
		4: Config4,
		5: Config5,
		6: Config6,
	}

	cars := vsetky[index]

	all_cars := Stav{Auta:[]Auto{}}
	for i := 0; i < len(cars); i++ {
		Auto := Auto{
			X:          uint8(cars[i][3].(int)),
			Y:          uint8(cars[i][2].(int)),
			Farba:      cars[i][0].(string),
			Dlzka:      uint8(cars[i][1].(int)),
		}
		if cars[i][4].(int32) == 'h' {
			Auto.Smer = true
		} else {
			Auto.Smer = false
		}
		all_cars.Auta = append(all_cars.Auta, Auto)
	}
	vypisMapu(vytvorMapuZAut(all_cars.Auta))
	return all_cars
}

func VypisKroky(cars Stav) {
	kroky := cars.Kroky
	for i := 0; i < len(kroky); i+=3 {

		switch kroky[i] {
		case 1:
			fmt.Print("HORE(")
		case 2:
			fmt.Print("VPRAVO(")
		case 3:
			fmt.Print("DOLE(")
		case 4:
			fmt.Print("VLAVO(")
		}
		fmt.Print(cars.Auta[kroky[i+1]].Farba, ",", kroky[i+2], ")\n")

	}
}

func vypisLokacieAut(stav Stav){
	fmt.Print("(")
	for i := 0; i < len(stav.Auta); i++ {
		fmt.Print("(", stav.Auta[i].Farba," ", stav.Auta[i].Dlzka, stav.Auta[i].Y, stav.Auta[i].X)
		if stav.Auta[i].Smer {
			fmt.Print(" h)")
		} else {
			fmt.Print(" v)")
		}
	}
	fmt.Println(")")
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for{
		fmt.Println("Test 1 - 15 krokov")
		fmt.Println("Test 2 - 8 krokov")
		fmt.Println("Test 3 - 6 krokov")
		fmt.Println("Test 4 - 4 kroky")
		fmt.Println("Test 5 - 2 kroky")
		fmt.Println("Test 6 - Nie je riešenie")
		fmt.Println("Zadajte číslo testovacieho vstupu")
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		i, err := strconv.Atoi(text)
		if err != nil {
			fmt.Println("Konfiguračný súbor neexistuje")
		}

		cars := loadConfig(i)
		vypisLokacieAut(cars)
		start := time.Now()
		finalny_stav := najdiRiesenie(cars, cars.Auta[0].Farba)
		if finalny_stav == nil {
			fmt.Println("Nebolo nájdené žiadne riešenie..")
			duration := time.Since(start)
			fmt.Println("Vypočítané za: ", duration)
			continue
		}
		duration := time.Since(start)
		VypisKroky(*finalny_stav)
		fmt.Println("Vypočítané za: ", duration)
		vypisMapu(finalny_stav.Mapa)
		vypisLokacieAut(*finalny_stav)
	}


}
