/* This program demonstrates recursive functions. */

/* return the factorial of i */
int factr(int i)
{
  if(i<2) {
    return 1;
  }
  else {
     return i * factr(i-1);
  }
}

int main()
{
  print("Factorial of 10 is: ");
  print(factr(10));

  return 0;
}
