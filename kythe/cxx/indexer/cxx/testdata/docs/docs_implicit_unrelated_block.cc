// We don't associate unrelated implicit BCPL-style comments.

void f() {
  if (true) return;  // unrelated
  int x = 1;
}

//- VarX.node/kind variable
//- VarX named vname("x:1:0:f#n",_,_,_,_)
//- !{VarXD documents VarX}
