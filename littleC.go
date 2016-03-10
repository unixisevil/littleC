package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

type varType struct {
	name string
	typ  int
	val  int
}

type funcType struct {
	name    string
	rettype int
	pos     int
}

//show local variable stack for debug
func (m *machine) showLvars() {
	for _, v := range m.lvars {
		log.Printf("got %s = %d, type = %d\n", v.name, v.val, v.typ)
	}
}

//handle  five builtin functions
func getch(m *machine) int {
	ch := 0
	fmt.Scanf("%c", &ch)
	for m.LA(1) != RP {
		m.consume()
	}
	m.consume()
	return ch
}
func getnum(m *machine) int {
	num := 0
	fmt.Scanf("%d", &num)
	for m.LA(1) != RP {
		m.consume()
	}
	m.consume()
	return num
}
func putch(m *machine) int {
	ch := 0
	if m.LT(1).Text != "putch" {
		panic("expect 'putch' func name")
	} else {
		m.consume()
	}
	m.expr(&ch)
	fmt.Printf("%c", ch)
	return ch
}
func puts(m *machine) int {
	if m.LT(1).Text != "puts" {
		panic("expect 'puts' func name")
	}
	m.consumeN(2)
	if m.LA(1) != STR {
		panic("expect string")
	}
	fmt.Println(m.LT(1).Text[1 : len(m.LT(1).Text)-1])
	m.consume()
	if m.LA(1) != RP {
		panic("expect right paren in puts func")
	}
	m.consume()
	return 0
}
func print(m *machine) int {
	i := 0
	if m.LT(1).Text != "print" {
		panic("expect 'print' func name")
	}
	m.consume()
	if m.LA(2) == STR {
		m.consume()
	}
	if m.LA(1) == STR {
		fmt.Println(m.LT(1).Text[1 : len(m.LT(1).Text)-1])
		m.consume()
		if m.LA(1) != RP {
			panic("expect right paren in print func")
		} else {
			m.consume()
		}
	} else {
		m.expr(&i)
		fmt.Printf("%d ", i)
	}
	return 0
}

//show lookahead token buffer for debug
func (m *machine) showTokenBuffer() {
	for _, t := range m.lookahead {
		log.Printf("%q ", t.Text)
	}
	log.Printf("%d\n", m.p)
}

//process global variable declaration
func (m *machine) gvarDecl() {
	typ := m.LA(1)
	m.consume()
	for {
		m.gvarmap[m.LT(1).Text] = &varType{
			name: m.LT(1).Text,
			typ:  typ,
		}
		if m.LA(2) != COMMA {
			break
		} else {
			m.consumeN(2)
		}
	}
	if m.LA(2) == SEMI {
		m.consumeN(2)
	} else {
		panic("expect semicolon in gvarDecl func")
	}
}

//process local variable declaration
func (m *machine) lvarDecl() {
	typ := m.LA(1)
	m.consume()
	for {
		v := &varType{
			name: m.LT(1).Text,
			typ:  typ,
		}
		m.lvars = append(m.lvars, v)
		if m.LA(2) != COMMA {
			break
		} else {
			m.consumeN(2)
		}
	}
	if m.LA(2) == SEMI {
		m.consumeN(2)
	} else {
		panic("expect semicolon in lvarDecl func")
	}
}

//find location of all functions, save global variable in map
func (m *machine) preScan() {
	brace := 0
	for {
		if m.LA(1) == EOF {
			break
		}
		if m.LA(1) == Int || m.LA(1) == Char {
			if m.LA(2) == VAR && m.LA(3) == LP {
				m.funcmap[m.LT(2).Text] = &funcType{
					name:    m.LT(2).Text,
					rettype: m.LA(1),
					pos:     m.LT(1).Pos,
				}
				m.consumeN(3)
				for m.LA(1) != LB {
					m.consume()
				}
			} else if m.LA(2) == VAR && m.LA(3) != LP {
				m.gvarDecl()
			}
		} else if m.LA(1) == LB {
			m.consume()
			brace++
		}
		for brace > 0 {
			switch m.LA(1) {
			case LB:
				brace++
			case RB:
				brace--
			}
			m.consume()
		}
	}
}

//handle  function  call
func (m *machine) funcCall(funcName string) {
	if f, ok := m.funcmap[funcName]; !ok {
		panic(fmt.Sprintf("function %s undefine", funcName))
	} else {
		if funcName == "main" {
			m.jump(f.pos)
			for m.LA(1) != LB {
				m.consume()
			}
			m.execBlock()
			return
		}
		lvarPos := len(m.lvars) - 1
		//eat  func name
		m.consume()
		m.getArgs()
		retPos := m.LT(1).Pos
		if lvarPos >= 0 {
			m.funcPush(lvarPos)
		}
		m.jump(f.pos)
		m.getParams()
		m.execBlock()
		m.jump(retPos)
		if lvarPos >= 0 {
			m.lvars = m.lvars[:m.funcPop()+1]
		} else {
			m.lvars = m.lvars[:0]
		}
	}
}

//lexer jump position 'pos', refill token buffer
func (m *machine) jump(pos int) {
	m.input.jump(pos)
	for i := 1; i <= m.k; i++ {
		m.consume()
	}
}

//maintain top position of local variable stack
func (m *machine) funcPush(i int) {
	m.callstack = append(m.callstack, i)
}
func (m *machine) funcPop() int {
	i := m.callstack[len(m.callstack)-1]
	m.callstack = m.callstack[:len(m.callstack)-1]
	return i
}

//handle  function  return
func (m *machine) funcRet() {
	//eat  'return' token
	m.consume()
	val := 0
	m.expr(&val)
	m.consume()
	m.retVal = val
}

func (m *machine) assignVar(name string, val int) {
	low := 0
	if len(m.callstack) != 0 {
		low = m.callstack[len(m.callstack)-1] + 1
	}
	for i := len(m.lvars) - 1; i >= low; i-- {
		if name == m.lvars[i].name {
			m.lvars[i].val = val
			return
		}
	}
	for k, v := range m.gvarmap {
		if k == name {
			v.val = val
			return
		}
	}
	panic("not find variable: " + name)
}

func (m *machine) findVar(name string) int {
	low := 0
	if len(m.callstack) != 0 {
		low = m.callstack[len(m.callstack)-1] + 1
	}
	for i := len(m.lvars) - 1; i >= low; i-- {
		if name == m.lvars[i].name {
			return m.lvars[i].val
		}
	}
	for k, v := range m.gvarmap {
		if k == name {
			return v.val
		}
	}
	panic("not find variable: " + name)
}
func (m *machine) isVar(name string) bool {
	low := 0
	if len(m.callstack) != 0 {
		low = m.callstack[len(m.callstack)-1] + 1
	}
	for i := len(m.lvars) - 1; i >= low; i-- {
		if name == m.lvars[i].name {
			return true
		}
	}
	for k, _ := range m.gvarmap {
		if k == name {
			return true
		}
	}
	return false
}

//handle  {}  code  block
func (m *machine) execBlock() {
	value := 0
	block := false
	//log.Println("entering execBlock()")
	for {
		switch m.LA(1) {
		case VAR:
			m.expr(&value)
			if m.LA(1) != SEMI {
				panic("expect semicolon after expression")
			}
			m.consume()
		case LB:
			m.consume()
			block = true
		case RB:
			m.consume()
			//log.Println("exiting execBlock()")
			return
		case Char, Int:
			m.lvarDecl()
		case Return:
			m.funcRet()
			return
		case If:
			m.execIf()
		case Else:
			m.findEob()
		case While:
			m.execWhile()
		case Do:
			m.execDo()
		case For:
			m.execFor()
		case EOF:
			os.Exit(0)
		default:
			m.consume()
		}
		if !block {
			break
		}
	}
}

//handle  function call  arguments
func (m *machine) getArgs() {
	if m.LA(1) != LP {
		panic("expect '(' start argument list")
	}
	//fast path for no args func
	if m.LA(2) == RP {
		m.consumeN(2)
		return
	}
	//eat  '(' token
	m.consume()
	value := 0
	temps := []int{}
	for {
		m.expr(&value)
		temps = append(temps, value)
		if m.LA(1) != COMMA {
			break
		} else {
			m.consume()
		}
	}
	if m.LA(1) != RP {
		panic("expect ')' ends argument list")
	}
	m.consume()
	for i := len(temps) - 1; i >= 0; i-- {
		m.lvars = append(m.lvars, &varType{
			val: temps[i],
		})
	}
}

//handle function definition parameters
func (m *machine) getParams() {
	//skip  type funcname etc
	for m.LA(1) != LP {
		m.consume()
	}
	//fast path for no args func
	if m.LA(1) == LP && m.LA(2) == RP {
		m.consumeN(2)
		return
	}
	//eat  '(' token
	m.consume()
	i := len(m.lvars) - 1
	for i >= 0 {
		if m.LA(1) == RP {
			break
		}
		if m.LA(1) != Char && m.LA(1) != Int {
			panic("in function definition param list  type expected")
		}
		m.lvars[i].typ = m.LA(1)
		m.consume()
		m.lvars[i].name = m.LT(1).Text
		m.consume()
		i--
		if m.LA(1) != COMMA {
			break
		} else {
			m.consume()
		}
	}
	if m.LA(1) != RP {
		panic("expect right paren after params list")
	}
	m.consume()
}

//skip  {} code block
func (m *machine) findEob() {
	brace := 0
	for {
		switch m.LA(1) {
		case LB:
			brace++
		case RB:
			brace--
		}
		m.consume()
		if brace == 0 {
			break
		}
	}
}

//handle  if stmt
func (m *machine) execIf() {
	//eat 'if' token
	m.consume()
	cond := 0
	m.expr(&cond)
	if cond != 0 {
		m.execBlock()
	} else {
		m.findEob()
		if m.LA(1) != Else {
			return
		}
		//eat 'else' token
		m.consume()
		m.execBlock()
	}
}

//handle  while stmt
func (m *machine) execWhile() {
	cond := 0
	whilePos := m.LT(1).Pos
	//eat 'while' token
	m.consume()
	m.expr(&cond)
	if cond != 0 {
		m.execBlock()
	} else {
		m.findEob()
		return
	}
	m.jump(whilePos)
}

//handle  do stmt
func (m *machine) execDo() {
	cond := 0
	doPos := m.LT(1).Pos
	//eat  'do' token
	m.consume()
	m.execBlock()
	if m.LA(1) != While {
		panic("expect 'while' token")
	}
	//eat 'while' token
	m.consume()
	m.expr(&cond)
	if cond != 0 {
		m.jump(doPos)
	}
}

//handle for stmt
func (m *machine) execFor() {
	cond := 0
	condPos, postPos := 0, 0
	if m.LA(1) != For {
		panic("expect 'for' token")
	}
	//eat  'for', '(' token
	m.consumeN(2)
	//compute  init stmt
	m.expr(&cond)
	if m.LA(1) != SEMI {
		panic("expect semicolon after init stmt")
	}
	//eat  ';' token
	condPos = m.LT(1).Pos + 1
	m.consume()
	for {
		m.expr(&cond)
		if m.LA(1) != SEMI {
			panic("expect semicolon after cond stmt")
		}
		//eat  ';' token
		postPos = m.LT(1).Pos + 1
		m.consume()
		//have seen one '(' after 'for'
		paren := 1
		for paren != 0 {
			switch m.LA(1) {
			case LP:
				paren++
			case RP:
				paren--
			}
			m.consume()
		}
		if cond != 0 {
			m.execBlock()
		} else {
			m.findEob()
			return
		}
		m.jump(postPos)
		//compute post stmt
		m.expr(&cond)
		if m.LA(1) != RP {
			panic("after compute post stmt, expect ')'")
		}
		m.consume()
		m.jump(condPos)
	}
}

type machine struct {
	input     *lexer
	retVal    int
	callstack []int
	lvars     []*varType
	gvarmap   map[string]*varType
	funcmap   map[string]*funcType
	env       map[string]int
	lookahead []*Token
	k, p      int
}

func newMachine(input string, k int) *machine {
	m := &machine{
		input:     newLexer(input),
		gvarmap:   make(map[string]*varType),
		funcmap:   make(map[string]*funcType),
		env:       make(map[string]int),
		lookahead: make([]*Token, k),
		k:         k,
	}
	for i := 1; i <= k; i++ {
		m.consume()
	}
	return m
}

//load lookahead i  token object
func (m *machine) LT(i int) *Token {
	return m.lookahead[(m.p+i-1)%m.k]
}

//lookahead i
func (m *machine) LA(i int) int {
	return m.LT(i).Type
}

//eat current token, fill one new token from lexer
func (m *machine) consume() {
	t := m.input.nextToken()
	if t.Type == ERR {
		panic("find error when lexing " + t.Text)
	}
	m.lookahead[m.p] = t
	m.p = (m.p + 1) % m.k
}

func (m *machine) consumeN(n int) {
	for i := 0; i < n; i++ {
		m.consume()
	}
}
func (m *machine) expr(val *int) {
	//log.Println("enter expr() ", m.LT(1).Text)
	if tok := m.LA(1); tok == EOF {
		return
	} else if tok == SEMI {
		*val = 0
		return
	}
	m.assign(val)
}
func (m *machine) assign(val *int) {
	//log.Println("enter assign() ", m.LT(1).Text)
	if m.LA(1) == VAR && m.LA(2) == ASSIGN {
		tok := m.LT(1)
		m.consumeN(2)
		m.assign(val)
		m.assignVar(tok.Text, *val)
		return
	}
	m.relOp(val)
}
func (m *machine) relOp(val *int) {
	//log.Println("enter relop() ", m.LT(1).Text)
	bool2int := func(res bool) int {
		if res {
			return 1
		} else {
			return 0
		}
	}
	temp := 0
	m.addSub(val)
	if tok := m.LA(1); tok > STR && tok < Keyword {
		m.consume()
		m.addSub(&temp)
		switch tok {
		case LT:
			*val = bool2int(*val < temp)
		case LE:
			*val = bool2int(*val <= temp)
		case GT:
			*val = bool2int(*val > temp)
		case GE:
			*val = bool2int(*val >= temp)
		case EQ:
			*val = bool2int(*val == temp)
		case NE:
			*val = bool2int(*val != temp)
		}
	}
}
func (m *machine) addSub(val *int) {
	//log.Println("enter addsub() ", m.LT(1).Text)
	temp := 0
	typ := 0
	m.mulDiv(val)
	for m.LA(1) == PLUS || m.LA(1) == MINUS {
		typ = m.LA(1)
		m.consume()
		m.mulDiv(&temp)
		switch typ {
		case PLUS:
			*val = *val + temp
		case MINUS:
			*val = *val - temp
		}
	}
}
func (m *machine) mulDiv(val *int) {
	//log.Println("enter mulDiv() ", m.LT(1).Text)
	temp := 0
	typ := 0
	m.pow(val)
	for m.LA(1) == MUL || m.LA(1) == DIV || m.LA(1) == REM {
		typ = m.LA(1)
		m.consume()
		m.pow(&temp)
		switch typ {
		case MUL:
			*val = *val * temp
		case DIV:
			if temp == 0 {
				panic("div by zero")
			} else {
				*val = *val / temp
			}
		case REM:
			*val = *val % temp
		}
	}
}

func (m *machine) pow(val *int) {
	//log.Println("enter pow() ", m.LT(1).Text)
	temp, ex := 0, 0
	m.unary(val)
	if m.LA(1) == EXP {
		m.consume()
		m.pow(&temp)
		ex = *val
		if temp == 0 {
			*val = 1
			return
		}
		for t := temp - 1; t > 0; t-- {
			*val = *val * ex
		}
	}
}
func (m *machine) unary(val *int) {
	//log.Println("enter unary() ", m.LT(1).Text)
	typ := 0
	if m.LA(1) == PLUS || m.LA(1) == MINUS {
		typ = m.LA(1)
		m.consume()
	}
	m.paren(val)
	if typ == MINUS {
		*val = -(*val)
	}
}
func (m *machine) paren(val *int) {
	//log.Println("enter paren() ", m.LT(1).Text)
	if m.LA(1) == LP {
		m.consume()
		m.assign(val)
		if m.LA(1) != RP {
			panic("unblanced ()")
		}
		m.consume()
	} else {
		m.atom(val)
	}
}

func (m *machine) atom(val *int) {
	//log.Println("enter atom() ", m.LT(1).Text)
	switch m.LA(1) {
	case VAR:
		if f, ok := builtinFuncs[m.LT(1).Text]; ok {
			*val = f(m)
		} else if _, ok = m.funcmap[m.LT(1).Text]; ok {
			m.funcCall(m.LT(1).Text)
			*val = m.retVal
		} else {
			*val = m.findVar(m.LT(1).Text)
			m.consume()
		}
	case NUM:
		*val, _ = strconv.Atoi(m.LT(1).Text)
		m.consume()
	case CHAR:
		*val = int(m.LT(1).Text[1])
		m.consume()
	default:
		panic(fmt.Sprintf("unexpected token :%q", m.LT(1).Text))
	}
}

func init() {
	builtinFuncs = map[string]func(*machine) int{
		"getch":  getch,
		"putch":  putch,
		"puts":   puts,
		"print":  print,
		"getnum": getnum,
	}
}

var builtinFuncs = map[string]func(*machine) int{}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s  <filename.c>\n", os.Args[0])
		os.Exit(1)
	}
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("read file %s ,error : %s\n", os.Args[1], err)
		os.Exit(1)
	}
	defer func() {
		if e := recover(); e != nil {
			fmt.Println(e)
		}
	}()
	m := newMachine(string(b), 3)
	m.preScan()
	m.funcCall("main")
}
