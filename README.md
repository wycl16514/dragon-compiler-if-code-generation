本节我们要完成的目的是将下面带有if语句的代码转换为中间代码：
```
{int x; int y; int z;
		        x = 1;
				y = 2;
				if (y > x) {
					z = 2;
				}
				z = 3;
	}
```
完成本节代码后，上面代码会被转译成如下中间代码：
```
L1:
        x = 1
L3:
        y = 2
L4:
        iffalse y > x goto L5
L6:
        z = 2
L5:
        z = 3
L2:
```
注意看L4, L5, L6三个地方的分支，中间代码有一个指令叫iffalse，后面跟着一个表达式，如果表达式结果能转换为false，那么goto语句就产生作用，跳转到它对应的语句，如果表达式结果为true，那么控制流直接跳转到L4下面的语句。

我们先看看if语句对应的语法规则表达式：
```
stmt -> if "(" bool ")" stmt
bool -> expr rel expr
rel -> LE | LESS_OPERATOR | GE | GREATER_OPERATOR | NE | EQ
```
if语句的规则是，在关键字if后面必须跟着左括号，然后对应bool表达式，它实际上是两个算术表达式进行比较操作，也就是两个表达式之间对应这<, <=, >, >=, !=, ==等比较操作符，两个表达式比较完后就会得出true或是false的结果，然后使用iffalse指令在表达式比较结果上对执行流进行操作。

因此我们实现的方法是，在遇到if语句前先给他一个跳转标签，也就是前面例子中的L4,然后if条件成立时对应的语句集合其实是一个stmt，所有语句对应一个标签，也就是L5。由于if语句后面会跟着一个左大括号，里面对应这如果判断条件成立就要执行的代码，于是对应右大括号后面的语句就是if判断条件不成立时要执行的代码，那么这些代码对应的跳转标签就紧接着L5，也就是上面例子中的L6。因此本节难点在于：1，为if语句生成对应代码，由于我们要由浅入深，因此本节if对应判断条件就是两个ID对象，或是ID和Constant常量对象比较，后面我们还会加上&&和||这种运算符。2，如何决定跳转的标签号。这些逻辑不好用言语表述，还是得在代码实现和调试中更好理解。

接下来我看代码实现，首先要修改一下ExprInterface的接口:
```
type ExprInterface interface {
	NodeInterface
	Gen() ExprInterface
	Reduce() ExprInterface
	Type() *Type
	ToString() string
	//新增两个接口
	Jumping(t uint32, f uint32)
	EmitJumps(test string, t uint32, f uint32)
}
```
这里新增两个接口分别为Jumping和EmitJumps，它们用来设置if, if..else..,for, while, do..while等控制语句的跳转，由于接口修改了，因此任何实现它的实例都得修改，我们下面只显示正要的修改，其他修改他家可以直接下载代码查看，代码下载地址我在末尾给出。首先要修改的是expr.go:
```
func (e *Expr) Jumping(t uint32, f uint32) {
	e.EmitJumps(e.ToString(), t, f)
}

func (e * Expr) EmitJumps(test string, t uint32, f uint32) {
	if t != 0 && f != 0 {
		e.Emit("if " + test + " got L" + strconv.Itoa(int(t)))
		e.Emit("goto L" + strconv.Itoa(int(f)))
	} else if t != 0 {
		e.Emit("if " + test + " goto L" + strconv.Itoa(int(t)))
	} else if f != 0 {
		e.Emit("iffalse " + test + " goto L" + strconv.Itoa(int(f)))
	}
}
```
Jumping接收两个参数t,f，他们其实对应true和false，然后调用EmitJumps，后者根据传入的t,f值来输出跳转代码，如果t等于0或者是f等于0，那意味着不用输出对应的跳转代码。在上面代码中我们目前需要关系的是：
```
e.Emit("ifalse " + test + "goto L" + strconv.Itoa(int(f))) 
```
这条语句就输出了我们前面例子中对应的：
```
 iffalse y > x goto L5
```
原理我们算术表达式使用"+"或者"-"进行连接，现在表达式之间也能用">", "<", "<="等比较符号连接，因此我们也要针对这种情况创建一个实现ExprInterface接口的实例，因此我们创建文件rel.g实现代码如下：
```
package inter

import (
	"lexer"
)

type Rel struct {
	logic *Logic
	expr1 ExprInterface
	expr2 ExprInterface
	token *lexer.Token
}

func relCheckType(p1 *Type, p2 *Type) *Type {
	if p1.Lexeme == p2.Lexeme {
		return NewType("bool", lexer.BASIC, 1)
	}

	return nil
}

func NewRel(line uint32, token *lexer.Token,
	expr1 ExprInterface, expr2 ExprInterface) *Rel {
	return &Rel{
		logic: NewLogic(line, token, expr1, expr2, relCheckType),
		expr1: expr1,
		expr2: expr2,
		token: token,
	}
}

func (r *Rel) Errors(s string) error {
	return r.logic.Errors(s)
}

func (r *Rel) NewLabel() uint32 {
	return r.logic.NewLabel()
}

func (r *Rel) EmitLabel(l uint32) {
	r.logic.EmitLabel(l)
}

func (r *Rel) Emit(code string) {
	r.logic.Emit(code)
}

func (r *Rel) Gen() ExprInterface {
	return r.logic.Gen()
}

func (r *Rel) Reduce() ExprInterface {
	return r
}

func (r *Rel) Type() *Type {
	return r.logic.Type()
}

func (r *Rel) ToString() string {
	return r.logic.ToString()
}

func (r *Rel) Jumping(t uint32, f uint32) {
	expr1 := r.expr1.Reduce()
	expr2 := r.expr2.Reduce()
	test := expr1.ToString() + " " + r.token.ToString() + " " + expr2.ToString()
	r.EmitJumps(test, t, f)
}

func (r *Rel) EmitJumps(test string, t uint32, l uint32) {
	r.logic.EmitJumps(test, t, l)
}

```
上面代码中，构造函数NewRel接收三个重要参数，分别是token, expr1, expr2,后面两个就是对应要比较的算术表达式，token对应衔接两个表达式的比较符号，也就是">","<","<="这些。这里需要确定两个表达式expr1,expr2属于相同类型，如果一个类型是数值，另一个类型是字符串，那么它们在逻辑上就没有可比性，于是代码中有了relCheckType函数，它判断两个表达式的类型必须一样，构造rel节点才能合法。

其实不同类型也能比较，例如int和float应该能相互比较，只不过为了简单起见，我们暂时不做考虑。我能还需要关系Jumping的实现，它分别调用了两个表达式的Reduce接口，如果表达式是复杂类型，例如 (a+b) > (c+d)这种，那么expr1对应a+b，调用它的Reduce后，根据前面我们的实现，编译器会将a+b的结果赋值给一个临时寄存器，然后用该寄存器来表示它，也就是a+b会先转译成：
```
t1 = a + b
```
同理c+d会被转译成:
```
t2 = c + d
```
最后代码会生成中间指类似如下：
```
iffalse t1 > t2 goto L5
```
上面代码中用到一个logic对象，它的作用在后面我们实现&&,||这种连接符时才有用，因此这里我们先把它的代码贴出来，不过暂时不用理解它，因为它的我们本节的影响不大，logic.go的内容如下：
```
package inter

import (
	"errors"
	"lexer"
	"strconv"
)

/*
实现or, and , !等操作
*/

type Logic struct {
	expr      ExprInterface
	token     *lexer.Token
	expr1     ExprInterface
	expr2     ExprInterface
	expr_type *Type
	line      uint32
}

type CheckType func(type1 *Type, type2 *Type) *Type

func logicCheckType(type1 *Type, type2 *Type) *Type {

	if type1.Lexeme == "bool" && type2.Lexeme == "bool" {
		return type1
	}

	return nil
}

func NewLogic(line uint32, token *lexer.Token,
	expr1 ExprInterface, expr2 ExprInterface, checkType CheckType) *Logic {
	expr_type := checkType(expr1.Type(), expr2.Type())
	if expr_type == nil {
		err := errors.New("type error")
		panic(err)
	}

	return &Logic{
		expr:      NewExpr(line, token, expr_type),
		token:     token,
		expr1:     expr1,
		expr2:     expr2,
		expr_type: expr_type,
		line:      line,
	}
}

func (l *Logic) Errors(s string) error {
	return l.expr.Errors(s)
}

func (l *Logic) NewLabel() uint32 {
	return l.expr.NewLabel()
}

func (l *Logic) EmitLabel(label uint32) {
	l.expr.EmitLabel(label)
}

func (l *Logic) Emit(code string) {
	l.expr.Emit(code)
}

func (l *Logic) Type() *Type {
	return l.expr_type
}

func (l *Logic) Gen() ExprInterface {
	f := l.NewLabel()
	a := l.NewLabel()
	temp := NewTemp(l.line, l.expr_type)
	l.Jumping(0, f)
	l.Emit(temp.ToString() + " = true")
	l.Emit("goto L" + strconv.Itoa(int(a)))
	l.EmitLabel(f)
	l.Emit(temp.ToString() + "=false")
	l.EmitLabel(a)
	return temp
}

func (l *Logic) Reduce() ExprInterface {
	return l
}

func (l *Logic) ToString() string {
	return l.expr1.ToString() + " " + l.token.ToString() + " " + l.expr2.ToString()
}

func (l *Logic) Jumping(t uint32, f uint32) {
	l.expr.Jumping(t, f)
}

func (l *Logic) EmitJumps(test string, t uint32, f uint32) {
	l.expr.EmitJumps(test, t, f)
}

```
从代码可以看到，rel.EmitJumps虽然调用了logic的EmitJumps，但本质上都是调用expr.EmitJumps，因此logci对rel没有产生实质性影响。现在我们回到语法解析，增加其对if语句的解析，首先我们要创建一个继承了StmtInterface接口的If节点，它用来生成if语句对应的中间代码，其内容如下：
```
package inter

import (
	"errors"
)

type If struct {
	stmt    StmtInterface
	expr    ExprInterface
	if_stmt StmtInterface
}

func NewIf(line uint32, expr ExprInterface, if_stmt StmtInterface) *If {
	if expr.Type().Lexeme != "bool" {
		err := errors.New("bool type required in if")
		panic(err)
	}
	return &If{
		stmt:    NewStmt(line),
		expr:    expr,
		if_stmt: if_stmt,
	}
}

func (i *If) Errors(str string) error {
	return i.stmt.Errors(str)
}

func (i *If) NewLabel() uint32 {
	return i.stmt.NewLabel()
}

func (i *If) EmitLabel(label uint32) {
	i.stmt.EmitLabel(label)
}

func (i *If) Emit(code string) {
	i.stmt.Emit(code)
}

func (i *If) Gen(_ uint32, end uint32) {
	label := i.NewLabel()
	i.expr.Jumping(0, end)
	i.EmitLabel(label)
	i.if_stmt.Gen(label, end)
}

```
它的逻辑如下，由于if语句会对一个数学表达式进行判断，根据判断结果来执行if成立时对应的语句集合，因此If节点对应两个参数，一个是ExprInterface实例，它对应的就是if后面用于判断的表达式，另一个是StmtInterface实例，它对应if成立后要执行的语句。所以在它的Gen函数中，end对应如果if条件不成立所要执行的代码的跳转标签，它生成了一个label，对应的就是if判断成立时，所要执行语句块的标签。i.expr.Jumping是在解析if 后面表达式后，跳转到判断成立时对于语句的地址标签，i.if_stmt.Gen用于生成if判断条件成立后，大括号里面的语句。

我们再看看语法解析的过程，在list_parser.go中做如下修改：
```
func (s *SimpleParser) stmt() inter.StmtInterface {
	/*
		stmt -> if "(" bool ")" stmt
		bool -> expr rel expr
		rel -> LE | LESS_OPERATOR | GE | GREATER_OPERATOR | NE | EQ
	*/
	switch s.cur_tok.Tag {
	case lexer.IF:
		s.move_forward()
		err := s.matchLexeme("(")
		if err != nil {
			panic(err)
		}
		s.move_forward()
		x := s.bool()
		err = s.matchLexeme(")")
		if err != nil {
			panic(err)
		}
		s.move_forward() //越过 ）
		s.move_forward() //越过{
		s1 := s.stmt()
		err = s.matchLexeme("}")
		if err != nil {
			panic(err)
		}
		s.move_forward() //越过}
		return inter.NewIf(s.lexer.Line, x, s1)
	default:
		return s.expression()
	}

}

func (s *SimpleParser) bool() inter.ExprInterface {
	expr1 := s.expr()
	var tok *lexer.Token

	switch s.cur_tok.Tag {
	case lexer.LE:
	case lexer.LESS_OPERATOR:
		fallthrough
	case lexer.GE:
		fallthrough
	case lexer.GREATER_OPERATOR:
		fallthrough
	case lexer.NE:
		tok = lexer.NewTokenWithString(s.cur_tok.Tag, s.lexer.Lexeme)
	default:
		tok = nil
	}

	if tok == nil {
		panic("wrong operator in if")
	}
	s.move_forward()
	expr2 := s.expr()
	return inter.NewRel(s.lexer.Line, tok, expr1, expr2)
}
```
上面代码逻辑是，首先判断当前读到的是否为if标签，如果是，那么进入都if语句的解析流程，bool()解析if语句对应的判断条件，它首先解析比较符号左边的表达式，然后读取比较符号，然后解析右边的表达式，最后将左边表达式，比较符合，右边表达式合在一起形成一个Rel节点。Rel节点会结合到If节点里，If在Gen调用生成代码时，就会调用Rel节点生成判断表达式的代码。在语法解析中，产生If节点的时候，除了解析if后面的表达式，代码还通过stmt()来解析if大括号里面的代码，最终形成If节点后，它的Reduce函数也能为大括号里面的代码生成中间代码。



