// ============================================================================
// (c) Copyright 2003 Roman Dementiev, Lutz Kettner, Jens Mehnert, and Peter
// Sanders, MPI Informatik, Stuhlsatzenhausweg 85, 66123 Saarbrucken, Germany.
//
// This source code is described in the paper: 
// Engineering a Sorted List Data Structure for 32 Bit Keys.
// By Roman Dementiev, Lutz Kettner, Jens Mehnert and Peter Sanders. 
// In: Algorithm Engineering and Experiments (ALENEX'04), New Orleans, Jan 2004
// ----------------------------------------------------------------------------
//
// start.cpp
// $Revision: $
// $Date: $
//
// Main program: command line argument parsing and benchmark calls.
// ============================================================================

#include <cstdlib> 
#include <cassert>
#include <iostream>
#include "FTL.h"

using namespace std;

int main(int argc, char* argv[]) {  
    FTL away;

    if(argv[1]!=0){
        int val = atoi(argv[1]); // Auftragsnummer des tests

        if(argv[2]!=0){  
            int val2 = atoi(argv[2]);

            // Shorter running tests with one test-instance size
            // -------------------------------------------------
            if(val == 1)
                if(atoi(argv[2])>0)
                    away.locTest(val2,1000000,1);
           
            if(val == 2) // insert
                if(atoi(argv[2])>0)
                    away.insTest(val2,1);
      
            if(val == 3) // delete
                if(atoi(argv[2])>0)
                    away.delTest(val2,1);
      
            if(val == 4) // insert&Delete
                if(atoi(argv[2])>0)
                    away.insDelTest(val2,1);

            // Longer running tests with test-instances of incr. size
            // ------------------------------------------------------
            // Range is set at compile time in the Makefile
            int range = RANGE - 8;

            // Locate test
            if(val == 5){      
                int tmp = 256;    
                for(int i=0;i<range;i++){           
                    away.locTest(tmp,1000000,val2);
                    tmp=tmp+tmp;   
                }
            }

            // Delete test
            if(val == 6){      
                int tmp = 256;    
                for(int i=0;i<range;i++){           
                    away.delTest(tmp,val2);
                    tmp=tmp+tmp;   
                }
            }

            // Insert test
            if(val == 7){      
                int tmp = 256;    
                for(int i=0;i<range;i++){  
                    away.insTest(tmp,val2);
                    tmp=tmp+tmp;   
                }
            }

            /* Nachbar Insert test */
      
            if(val == 77){      
                int tmp = 256;    
                for(int i=0;i<range;i++){
                    away.consecutiveElements(tmp,val2);
                    tmp=tmp+tmp;   
                }
            }

            // Alternating between insert and delete
            if(val == 8){      
                int tmp = 256;    
                for(int i=0;i<range;i++){                
                    away.insDelTest(tmp,val2);
                    tmp=tmp+tmp;   
                }
            }

            // Alternating between insert and delete
            if(val == 10){      
                int tmp = 256;    
                for(int i=0;i<12;i++){                
                    away.insDelTest(val2,tmp);
                    tmp=tmp<<1;   
                }
            }
      
            // Delete test
            if(val == 11){      
                int tmp = 16;    
                for(int i=0;i<15;i++){           
                    away.delTest(val2,tmp);  
                    tmp=tmp<<1;   
                }
            }
      
            // Insert test
            if(val == 12){      
                int tmp = 16;    
                for(int i=0;i<15;i++){  
                    //FTL away2;              
                    away.insTest(val2,tmp);  
                    tmp=tmp<<1;   
                }
            }
        }
        // memory test
        if(val == 9){      
            int tmp = 256;    
            for(int i=0;i<12;i++){                
                away.insDelTest(tmp,0);
                tmp=tmp+tmp;   
            }
      
        }
    }
    return 0;	
}
