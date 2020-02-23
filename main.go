package main

import (
	"addressbook/addressbookpb"
	"bufio"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

type menuOption int

const (
	OptionCreatePerson menuOption = iota
	OptionShowPerson
	OptionUpdatePerson
	OptionDeletePerson
	OptionListAllPeople
	OptionQuit
)

const (
	addressBook = "addressbook.db"
	menu        = `What would you like to Do?
(0) - Create Person
	Enter the new person's name, email, and phone number
(1) - Show Person 
	Show a person's details using their contact ID #
(2) - Update Person
	Update a person's name, email, and phone number
(3) - Delete Person
	Delete a person using their contact ID #
(4) - List People 
	Show all saved contacts
(5) - Quit
	Quits the application
`
)

func main() {
	var ab addressbookpb.AddressBook
	if _, err := os.Stat(addressBook); os.IsNotExist(err) {
		err := writeToFile(addressBook, &ab)
		if err != nil {
			log.Fatalf("Couldn't write address book database: %v", err)
		}
	}
	err := readFromFile(addressBook, &ab)
	if err != nil {
		log.Fatalf("Couldn't read address book database: %v", err)
	}
	fmt.Println("Welcome to Address Book!")
	for {
		printMenu()
		readNextOption(&ab)
	}
}

func readNextOption(ab *addressbookpb.AddressBook) {
	in := readLine()
	option, err := strconv.Atoi(in)
	if err != nil {
		fmt.Println("Unknown option")
		return
	}
	handleNextOption(ab, option)
}

func readLine() string {
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		in := scanner.Text()
		return in
	}
	return ""
}

func handleNextOption(ab *addressbookpb.AddressBook, option int) {
	switch menuOption(option) {
	case OptionCreatePerson:
		if err := createPerson(ab); err != nil {
			fmt.Println(err)
		}
	case OptionShowPerson:
		if err := showPerson(ab); err != nil {
			fmt.Println(err)
		}
	case OptionUpdatePerson:
		if err := updatePerson(ab); err != nil {
			fmt.Println(err)
		}
	case OptionDeletePerson:
		if err := deletePerson(ab); err != nil {
			fmt.Println(err)
		}
	case OptionListAllPeople:
		err := listAllPeople(ab)
		if err != nil {
			fmt.Println(err)
		}
	case OptionQuit:
		err := writeToFile(addressBook, ab)
		if err != nil {
			log.Fatalf("Couldn't write address book database: %v", err)
		}
		fmt.Println("Goodbye!")
		os.Exit(0)
	default:
		fmt.Println("Unknown option. Try again.")
	}
}

func createPerson(ab *addressbookpb.AddressBook) error {
	var p addressbookpb.Person
	fmt.Print("Name: ")
	p.Name = readLine()

	fmt.Print("E-mail: ")
	p.Email = readLine()

	fmt.Print("Phone #: ")
	var phone addressbookpb.Person_PhoneNumber
	phone.Number = readLine()

	fmt.Print("Phone Type (Mobile - 0, Home - 1, Work - 2): ")
	in := readLine()

	pt, err := strconv.Atoi(in)
	if err != nil {
		return fmt.Errorf("can't create person: %v", err)
	}
	phone.Type = addressbookpb.Person_PhoneType(pt)

	p.Phones = append(p.GetPhones(), &phone)
	ab.People = append(ab.GetPeople(), &p)
	return nil
}

func showPerson(ab *addressbookpb.AddressBook) error {
	id, person, err := searchByID(ab)
	if err != nil {
		return err
	}
	printPerson(id, person)
	return nil
}

func searchByID(ab *addressbookpb.AddressBook) (int, *addressbookpb.Person, error) {
	fmt.Print("Contact ID #: ")
	in := readLine()
	id, err := strconv.Atoi(in)
	if err != nil {
		return 0, nil, fmt.Errorf("unknown contact ID %s: %v", in, err)
	}
	var person *addressbookpb.Person
	for i, p := range ab.GetPeople() {
		if i == id {
			person = p
			break
		}
	}
	if person == nil {
		return 0, nil, fmt.Errorf("couldn't find person with contact ID: %s", in)
	}
	return 0, person, nil
}

func updatePerson(ab *addressbookpb.AddressBook) error {
	_, p, err := searchByID(ab)
	if err != nil {
		return err
	}
	fmt.Printf("Name (%s): ", p.Name)
	if name := readLine(); strings.TrimSpace(name) != "" {
		p.Name = name
	}

	fmt.Printf("E-mail (%s): ", p.Email)
	if email := readLine(); strings.TrimSpace(email) != "" {
		p.Email = email
	}

	fmt.Printf("Phone # (%s): ", p.GetPhones()[0])
	phone := p.GetPhones()[0]
	if number := readLine(); strings.TrimSpace(number) != "" {
		phone.Number = number
	}

	fmt.Printf("Phone Type (Mobile - 0, Home - 1, Work - 2): ")
	if pt := readLine(); strings.TrimSpace(pt) != "" {
		phoneType, err := strconv.Atoi(pt)
		if err != nil {
			return fmt.Errorf("can't create person: %v", err)
		}
		phone.Type = addressbookpb.Person_PhoneType(phoneType)
	}
	return nil
}

func deletePerson(ab *addressbookpb.AddressBook) error {
	fmt.Print("Contact ID #: ")
	in := readLine()
	id, err := strconv.Atoi(in)
	if err != nil {
		return fmt.Errorf("couldn't delete person with contact ID %s: %v", in, err)
	}
	if id < len(ab.GetPeople())-1 {
		copy(ab.GetPeople()[id:], ab.GetPeople()[id+1:])
	}
	ab.GetPeople()[len(ab.GetPeople())-1] = nil
	ab.People = ab.GetPeople()[:len(ab.GetPeople())-1]
	fmt.Printf("%d has been deleted", id)
	return nil
}

func listAllPeople(ab *addressbookpb.AddressBook) error {
	fmt.Println("Contact Listing:")
	for i, person := range ab.GetPeople() {
		printPerson(i, person)
	}
	return nil
}

func printPerson(i int, person *addressbookpb.Person) {
	fmt.Printf("%d - %s\n", i, person.String())
}

func printMenu() {
	fmt.Print(menu)
}

func writeToFile(fname string, pb proto.Message) error {
	bytes, err := proto.Marshal(pb)
	if err != nil {
		return fmt.Errorf("can't serialize pb to bytes: %v", err)
	}
	if err = ioutil.WriteFile(fname, bytes, 0644); err != nil {
		return fmt.Errorf("can't write bytes to file: %v\n", err)
	}
	return nil
}

func readFromFile(fname string, pb proto.Message) error {
	bytes, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("couldn't read %s: %v", fname, err)
	}
	if err = proto.Unmarshal(bytes, pb); err != nil {
		return fmt.Errorf("failed to unmarshal bytes: %v", err)
	}
	return nil
}
