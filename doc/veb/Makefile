#  Makefile
#
#  Engineering a Sorted List Data Structure for 32 Bit Keys
#
#  (c) Copyright 2003
#      Roman Dementiev, Lutz Kettner, Jens Mehnert, and Peter Sanders
#      MPI Informatik, Stuhlsatzenhausweg 85, 66123 Saarbr\"ucken, Germany
#      [dementiev,kettner,sanders,snej]@mpi-sb.mpg.de

#  Installation:
#      set the following variables according to your local environment

# set the desired struct, test numbers and and ranges
STRUCT = STREE # the data structure that we test, set one of the following:
               # ORIG_STREE     # original untuned veb tree
               # STREE          # our veb tree
               # STLMAP         # STL map
               # SORTSEQ        # LEDA sorted sequence, implemented as skiplist
               # LEDA_STREE_DP  # LEDA veb impl., uses dynamic perfect hashing
               # LEDA_STREE     # LEDA veb impl., usus hashing with chaining
               # LEDA_AB_TREE   # LEDA ab-tree

TIMES  = 10    # we repeat the test TIMES often

RANGE  = 18    # we store 2^RANGE many elements

# Your local LEDA installation. Using the std::allocator, our implementation
# is independent of LEDA and could run alone.
LEDAROOT = /usr/local/LEDA-4.2.1/

#  Set one of the following preprocessor definitions to select a memory manager
# -DUSE_STD_ALLOCATOR        # std:: g++ STL allocator, single threaded
# -DUSE_STD_MT_ALLOCATOR     # std:: g++ STL allocator, multi threaded
# -DUSE_LEDA_ALLOCATOR       # leda:: default allocator, single threaded
# -DUSE_LEDA_BIG_ALLOCATOR   # leda:: allocator, increased size, our choice

#  In addition to allocators one can set this to get the LEDA_MEMORY macro
#  style of allocations where applicable.
# -DUSE_LEDA_MEMORY

# Set this macro if you are running this on a g++-3.0 or higher
# -DGCC3

# set the appropriate optimization and debug levels
CPPFLAGS = 	-DTEST_$(STRUCT)			\
		-DTIMES=$(TIMES)			\
		-DRANGE=$(RANGE)			\
		-I$(LEDAROOT)/incl			\
		-Wall					\
		-Wno-deprecated				\
		-O6 -DNDEBUG -DLEDA_CHECKING_OFF	\
		-DUSE_LEDA_BIG_ALLOCATOR 

# the required libraries from LEDA etc.
LDFLAGS  = 	-O6 -Wl,-R$(LEDAROOT) -L$(LEDAROOT) -lL -lm 

GCC = g++

# Probably nothing to change below this line

HEADER = 	Dlist.h		\
		FTL.h		\
		LPHash.h	\
		LVL1Tree.h	\
		LVL2Tree.h	\
		LVL3Tree.h	\
		Top1.h		\
		Top23.h		\
		allocator.h	\
		timer.h

start:  start.o
	$(GCC)  start.o -o start $(LDFLAGS)


start.o: start.cpp $(HEADER)
	$(GCC) $(CPPFLAGS) -c start.cpp

leda_eb_tree:
	$(GCC) $(CPPFLAGS) -c eb_tree.cpp

leda_dp_hash:
	$(GCC) $(CPPFLAGS) -c dp_hash.cpp

clean:
	find . -name "*.o" -exec rm "{}" \;


# some helper rules in running benchmarks and evaluating the results

# set the shift to a value between 0 and 31: the random 32 bit keys
# will be shifted to the right using this SHIFT value. This gives
# us denser set of keys that stress the level 2 and level 3 trees better.
SHIFT = 0

nachbar:
	rm -f insertN.eps
	rm -f insertN.log
	./start 77 $(SHIFT) > insertN.log
	gnuplot insertN.gnu

gnuall:
	gnuplot insertP.gnu
	gnuplot deleteP.gnu
	gnuplot locateP.gnu
	gnuplot insDelP.gnu

runall:
	rm -f insertP.eps
	rm -f insert.log
	./start 7 $(SHIFT) > insert.log

	rm -f deleteP.eps
	rm -f delete.log
	./start 6 $(SHIFT) > delete.log

	rm -f locateP.eps
	rm -f locate.log
	./start 5 $(SHIFT) > locate.log

	rm -f insDelP.eps
	rm -f insDel.log
	./start 8 $(SHIFT) > insDel.log
