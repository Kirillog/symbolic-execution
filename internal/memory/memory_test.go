package memory

import (
	"testing"

	"github.com/ebukreev/go-z3/z3"

	"symbolic-execution-course/internal/symbolic"
	"symbolic-execution-course/internal/translator"
)

func TestMemoryAliasing(t *testing.T) {
	heap := NewSymbolicMemory()
	foo := &symbolic.StructT{Name: "Foo", Fields: []symbolic.ExpressionType{symbolic.IntType}}
	first := &symbolic.Ref{Addr: symbolic.NewSymbolicVariable("foo1", symbolic.AddrType), ExprType: foo}
	second := &symbolic.Ref{Addr: symbolic.NewSymbolicVariable("foo2", symbolic.AddrType), ExprType: foo}
	two, five := symbolic.NewIntConstant(2), symbolic.NewIntConstant(5)
	heap.AssignField(second, 0, five)
	heap.AssignField(first, 0, two)
	got := heap.GetFieldValue(second, 0)

	for _, tc := range []struct {
		name     string
		relation symbolic.BinaryOperator
		want     symbolic.SymbolicExpression
	}{
		{"same pointer", symbolic.EQ, two},
		{"distinct pointers", symbolic.NE, five},
	} {
		t.Run(tc.name, func(t *testing.T) {
			conditions := []symbolic.SymbolicExpression{
				symbolic.NewBinaryOperation(first, second, tc.relation),
				symbolic.NewBinaryOperation(first, &symbolic.Ref{Addr: symbolic.NullAddr, ExprType: foo}, symbolic.NE),
				symbolic.NewBinaryOperation(second, &symbolic.Ref{Addr: symbolic.NullAddr, ExprType: foo}, symbolic.NE),
			}
			checkMemoryResult(t, got, tc.want, conditions...)
		})
	}
}

func TestMemoryPathConstraintMutability(t *testing.T) {
	heap := NewSymbolicMemory()
	ref := heap.Allocate(personType())
	age := symbolic.NewSymbolicVariable("age", symbolic.IntType)
	eighteen := symbolic.NewIntConstant(18)
	heap.AssignField(ref, 1, age)
	before := heap.GetFieldValue(ref, 1)
	condition := symbolic.NewBinaryOperation(before, eighteen, symbolic.NE)
	heap.AssignField(ref, 1, eighteen)
	checkMemoryResult(t, heap.GetFieldValue(ref, 1), eighteen, condition)
}

func TestMemoryArrayOfStructs(t *testing.T) {
	heap := NewSymbolicMemory()
	person := personType()
	people := heap.Allocate(symbolic.ArrayT{Elem: person})
	bob := heap.Allocate(person)
	name := &symbolic.StringConstant{Value: "Bob"}
	thirty := symbolic.NewIntConstant(30)
	thirtyFive := symbolic.NewIntConstant(35)
	heap.AssignField(bob, 0, name)
	heap.AssignField(bob, 1, thirty)
	heap.AssignToArray(people, 1, heap.Read(Root(bob)))
	heap.Write(Root(people).Element(symbolic.NewIntConstant(1)).Field(1), thirtyFive)

	array := heap.Read(Root(people)).(*ArrayValue)
	element := array.Element(symbolic.NewIntConstant(1)).(*StructValue)
	checkMemoryResult(t, element.Field(0), name)
	checkMemoryResult(t, element.Field(1), thirtyFive)
	checkMemoryResult(t, heap.GetFieldValue(bob, 1), thirty)
}

func TestMemoryStructBasic(t *testing.T) {
	heap := NewSymbolicMemory()
	person := heap.Allocate(personType())
	checkPerson(t, heap.Read(Root(person)).(*StructValue), "", 0, 0)
	heap.AssignField(person, 0, &symbolic.StringConstant{Value: "Alice"})
	heap.AssignField(person, 1, symbolic.NewIntConstant(25))
	heap.AssignField(person, 2, symbolic.NewIntConstant(1001))
	checkPerson(t, heap.Read(Root(person)).(*StructValue), "Alice", 25, 1001)
}

func TestMemoryStructModification(t *testing.T) {
	heap := NewSymbolicMemory()
	person := heap.Allocate(personType())
	name := symbolic.NewSymbolicVariable("name", symbolic.StringType)
	heap.AssignField(person, 0, name)
	heap.AssignField(person, 1, symbolic.NewIntConstant(25))
	heap.AssignField(person, 2, symbolic.NewIntConstant(1001))
	updated := heap.Allocate(person.ExprType)
	heap.Write(Root(updated), heap.Read(Root(person)))
	updateInt(&heap, Root(updated).Field(1), symbolic.ADD, 1)
	updateInt(&heap, Root(updated).Field(2), symbolic.MUL, 2)
	for _, tc := range []struct {
		ref     *symbolic.Ref
		age, id int64
	}{
		{person, 25, 1001},
		{updated, 26, 2002},
	} {
		nonEmpty := symbolic.NewBinaryOperation(name, &symbolic.StringConstant{}, symbolic.NE)
		checkMemoryResult(t, heap.GetFieldValue(tc.ref, 0), name, nonEmpty)
		checkMemoryResult(t, heap.GetFieldValue(tc.ref, 1), symbolic.NewIntConstant(tc.age))
		checkMemoryResult(t, heap.GetFieldValue(tc.ref, 2), symbolic.NewIntConstant(tc.id))
	}
}

func TestMemoryStructPointer(t *testing.T) {
	heap := NewSymbolicMemory()
	person := allocatePerson(&heap, personType(), "Bob", 30, 2002)
	updateInt(&heap, Root(person).Field(1), symbolic.ADD, 5)
	checkPerson(t, heap.Read(Root(person)).(*StructValue), "Bob", 35, 2002)
}

func TestMemoryStructPointerModification(t *testing.T) {
	heap := NewSymbolicMemory()
	person := allocatePerson(&heap, personType(), "Bob", 30, 2002)
	pointer := *person
	modifyPersonThroughPointer(&heap, &pointer)
	checkPerson(t, heap.Read(Root(person)).(*StructValue), "Bob", 40, 3002)
}

func TestMemoryNestedStructPointer(t *testing.T) {
	heap := NewSymbolicMemory()
	person := allocatePerson(&heap, personType(), "Bob", 30, 2002)
	updateInt(&heap, Root(person).Field(1), symbolic.ADD, 5)
	pointerSlot := heap.Allocate(person.Type())
	heap.Write(Root(pointerSlot), person)
	pointer := heap.Read(Root(pointerSlot)).(*symbolic.Ref)
	modifyPersonThroughPointer(&heap, pointer)
	checkPerson(t, heap.Read(Root(person)).(*StructValue), "Bob", 45, 3002)
}

func TestMemoryArrayFixed(t *testing.T) {
	heap := NewSymbolicMemory()
	array := heap.Allocate(symbolic.ArrayT{Elem: symbolic.IntType})
	checkMemoryResult(t, heap.GetFromArray(array, 4), symbolic.NewIntConstant(0))
	for i := range 5 {
		heap.AssignToArray(array, i, symbolic.NewIntConstant(int64(i*i)))
	}
	checkInts(t, &heap, array, 0, 1, 4, 9, 16)
}

func TestMemoryArrayModification(t *testing.T) {
	heap := NewSymbolicMemory()
	array := allocateInts(&heap, 0, 1, 4, 9, 16)
	updated := heap.Allocate(array.ExprType)
	heap.Write(Root(updated), heap.Read(Root(array)))
	for i := range int64(5) {
		updateInt(&heap, Root(updated).Element(symbolic.NewIntConstant(i)), symbolic.ADD, 1)
	}
	checkInts(t, &heap, array, 0, 1, 4, 9, 16)
	checkInts(t, &heap, updated, 1, 2, 5, 10, 17)
}

// The slice examples exercise the backing arrays through the memory API.
func TestMemorySliceCreation(t *testing.T) {
	heap := NewSymbolicMemory()
	backing := heap.Allocate(symbolic.ArrayT{Elem: symbolic.IntType})
	for i := range 5 {
		heap.AssignToArray(backing, i, symbolic.NewIntConstant(int64(i*2)))
	}
	checkInts(t, &heap, backing, 0, 2, 4, 6, 8)
}

func TestMemorySliceAppend(t *testing.T) {
	heap := NewSymbolicMemory()
	arrayType := symbolic.ArrayT{Elem: symbolic.IntType}
	var backing *symbolic.Ref
	for i := range 3 {
		grown := heap.Allocate(arrayType)
		if backing != nil {
			heap.Write(Root(grown), heap.Read(Root(backing)))
		}
		heap.AssignToArray(grown, i, symbolic.NewIntConstant(int64(i*10)))
		backing = grown
	}
	checkInts(t, &heap, backing, 0, 10, 20)
}

func TestMemorySliceModification(t *testing.T) {
	heap := NewSymbolicMemory()
	backing := allocateInts(&heap, 1, 2, 3)
	view := *backing
	for i := range int64(3) {
		updateInt(&heap, Root(&view).Element(symbolic.NewIntConstant(i)), symbolic.MUL, 2)
	}

	grown := heap.Allocate(backing.ExprType)
	heap.Write(Root(grown), heap.Read(Root(&view)))
	heap.AssignToArray(grown, 3, symbolic.NewIntConstant(999))
	checkInts(t, &heap, backing, 2, 4, 6, 0)
	checkInts(t, &heap, grown, 2, 4, 6, 999)
}

func TestMemoryStructWithArray(t *testing.T) {
	heap := NewSymbolicMemory()
	grades := allocateInts(&heap, 85, 90, 78, 92, 88)
	student := heap.Allocate(&symbolic.StructT{Name: "Student", Fields: []symbolic.ExpressionType{
		symbolic.StringType, grades.ExprType, symbolic.FloatType,
	}})
	name := &symbolic.StringConstant{Value: "Charlie"}
	heap.AssignField(student, 0, name)
	heap.AssignField(student, 1, heap.Read(Root(grades)))
	heap.AssignField(student, 2, symbolic.NewFloatConstant(float64(85+90+78+92+88)/5))
	heap.AssignToArray(grades, 0, symbolic.NewIntConstant(100))

	checkMemoryResult(t, heap.GetFieldValue(student, 0), name)
	storedGrades := heap.GetFieldValue(student, 1).(*ArrayValue)
	for i, grade := range []int64{85, 90, 78, 92, 88} {
		checkMemoryResult(t, storedGrades.Element(symbolic.NewIntConstant(int64(i))), symbolic.NewIntConstant(grade))
	}
	checkMemoryResult(t, heap.GetFieldValue(student, 2), symbolic.NewFloatConstant(86.6))
}

func TestMemoryNestedStructs(t *testing.T) {
	heap := NewSymbolicMemory()
	employee := allocateEmployee(&heap)
	value := heap.Read(Root(employee)).(*StructValue)
	checkPerson(t, value.Field(0).(*StructValue), "David", 35, 3003)
	address := value.Field(1).(*StructValue)
	checkMemoryResult(t, address.Field(0), &symbolic.StringConstant{Value: "Main St"})
	checkMemoryResult(t, address.Field(1), &symbolic.StringConstant{Value: "Boston"})
	checkMemoryResult(t, symbolic.NewBinaryOperation(address.Field(0), address.Field(1), symbolic.NE), symbolic.NewBoolConstant(true))
	checkMemoryResult(t, address.Field(2), symbolic.NewIntConstant(12345))
	checkMemoryResult(t, value.Field(2), symbolic.NewFloatConstant(75000))
}

func TestMemoryNestedStructModification(t *testing.T) {
	heap := NewSymbolicMemory()
	employee := allocateEmployee(&heap)
	root := Root(employee)
	updateInt(&heap, root.Field(0).Field(1), symbolic.ADD, 1)
	salary := symbolic.NewFloatConstant(1.1)
	heap.AssignField(employee, 2, symbolic.NewBinaryOperation(heap.GetFieldValue(employee, 2), salary, symbolic.MUL))
	heap.Write(root.Field(1).Field(2), symbolic.NewIntConstant(54321))

	checkPerson(t, heap.GetFieldValue(employee, 0).(*StructValue), "David", 36, 3003)
	address := heap.GetFieldValue(employee, 1).(*StructValue)
	checkMemoryResult(t, address.Field(0), &symbolic.StringConstant{Value: "Main St"})
	checkMemoryResult(t, address.Field(2), symbolic.NewIntConstant(54321))
	baseSalary := float64(75000)
	checkMemoryResult(t, heap.GetFieldValue(employee, 2), symbolic.NewFloatConstant(baseSalary*1.1))
}

func TestMemoryArrayOfStructsModification(t *testing.T) {
	heap := NewSymbolicMemory()
	person := personType()
	people := heap.Allocate(symbolic.ArrayT{Elem: person})
	names := []string{"Alice", "Bob", "Charlie"}
	ages := []int64{25, 30, 35}
	for i, name := range names {
		value := allocatePerson(&heap, person, name, ages[i], int64(i+1))
		heap.AssignToArray(people, i, heap.Read(Root(value)))
	}
	pointer := *people
	for i := range names {
		id := Root(&pointer).Element(symbolic.NewIntConstant(int64(i))).Field(2)
		updateInt(&heap, id, symbolic.MUL, 10)
	}
	for i, name := range names {
		checkPerson(t, heap.GetFromArray(people, i).(*StructValue), name, ages[i], int64(i+1)*10)
	}
}

func personType() *symbolic.StructT {
	return &symbolic.StructT{Name: "Person", Fields: []symbolic.ExpressionType{
		symbolic.StringType, symbolic.IntType, symbolic.IntType,
	}}
}

func allocatePerson(heap *SymbolicMemory, tpe *symbolic.StructT, name string, age, id int64) *symbolic.Ref {
	person := heap.Allocate(tpe)
	heap.AssignField(person, 0, &symbolic.StringConstant{Value: name})
	heap.AssignField(person, 1, symbolic.NewIntConstant(age))
	heap.AssignField(person, 2, symbolic.NewIntConstant(id))
	return person
}

func allocateEmployee(heap *SymbolicMemory) *symbolic.Ref {
	person := allocatePerson(heap, personType(), "David", 35, 3003)
	address := &symbolic.StructT{Name: "Address", Fields: []symbolic.ExpressionType{
		symbolic.StringType, symbolic.StringType, symbolic.IntType,
	}}
	employee := heap.Allocate(&symbolic.StructT{Name: "Employee", Fields: []symbolic.ExpressionType{
		person.ExprType, address, symbolic.FloatType,
	}})
	heap.AssignField(employee, 0, heap.Read(Root(person)))
	home := Root(employee).Field(1)
	heap.Write(home.Field(0), &symbolic.StringConstant{Value: "Main St"})
	heap.Write(home.Field(1), &symbolic.StringConstant{Value: "Boston"})
	heap.Write(home.Field(2), symbolic.NewIntConstant(12345))
	heap.AssignField(employee, 2, symbolic.NewFloatConstant(75000))
	return employee
}

func allocateInts(heap *SymbolicMemory, values ...int64) *symbolic.Ref {
	array := heap.Allocate(symbolic.ArrayT{Elem: symbolic.IntType})
	for i, value := range values {
		heap.AssignToArray(array, i, symbolic.NewIntConstant(value))
	}
	return array
}

func modifyPersonThroughPointer(heap *SymbolicMemory, person *symbolic.Ref) {
	updateInt(heap, Root(person).Field(1), symbolic.ADD, 10)
	updateInt(heap, Root(person).Field(2), symbolic.ADD, 1000)
}

func updateInt(heap *SymbolicMemory, address Address, op symbolic.BinaryOperator, operand int64) {
	heap.Write(address, symbolic.NewBinaryOperation(heap.Read(address), symbolic.NewIntConstant(operand), op))
}

func checkPerson(t *testing.T, person *StructValue, name string, age, id int64) {
	t.Helper()
	checkMemoryResult(t, person.Field(0), &symbolic.StringConstant{Value: name})
	checkMemoryResult(t, person.Field(1), symbolic.NewIntConstant(age))
	checkMemoryResult(t, person.Field(2), symbolic.NewIntConstant(id))
}

func checkInts(t *testing.T, heap *SymbolicMemory, array *symbolic.Ref, want ...int64) {
	t.Helper()
	for i, value := range want {
		checkMemoryResult(t, heap.GetFromArray(array, i), symbolic.NewIntConstant(value))
	}
}

func checkMemoryResult(t *testing.T, got, want symbolic.SymbolicExpression, conditions ...symbolic.SymbolicExpression) {
	t.Helper()
	tr := translator.NewIntTranslator()
	solver := z3.NewSolver(tr.GetContext().(*z3.Context))
	assert := func(expr symbolic.SymbolicExpression) {
		value, err := tr.TranslateExpression(expr)
		if err != nil {
			t.Fatal(err)
		}
		solver.Assert(value.(z3.Bool))
	}
	for _, condition := range conditions {
		assert(condition)
	}
	if sat, err := solver.Check(); err != nil || !sat {
		t.Fatalf("expected a satisfiable input, got %v (%v)", sat, err)
	}
	if got.Type() == symbolic.StringType {
		model := solver.Model()
		actual, err := tr.Evaluate(model, got)
		if err != nil {
			t.Fatal(err)
		}
		expected, known := want.(*symbolic.StringConstant)
		if !known {
			value, err := tr.Evaluate(model, want)
			if err != nil {
				t.Fatal(err)
			}
			expected = value.(*symbolic.StringConstant)
		}
		if actual.(*symbolic.StringConstant).Value != expected.Value {
			t.Fatalf("in model, got %s, want %s", actual, expected)
		}
	}
	assert(symbolic.NewBinaryOperation(got, want, symbolic.NE))
	if sat, err := solver.Check(); err != nil || sat {
		t.Fatalf("expected %s to equal %s, got %v (%v)", got, want, sat, err)
	}
}
