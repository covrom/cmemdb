
#include <stdlib.h>
#include <string.h>
#include <stdio.h>

#include "str_map.h"
#include "../src/ahtable.h"

/* Simple random string generation. */
void randstr(char* x, size_t len)
{
    x[len] = '\0';
    while (len > 0) {
        x[--len] = '\x20' + (rand() % ('\x7e' - '\x20' + 1));
    }
}


const size_t n = 100000;  // how many unique strings
const size_t m_low  = 50;  // minimum length of each string
const size_t m_high = 500; // maximum length of each string
const size_t k = 200000;  // number of insertions
char** xs;

ahtable_t* T;
str_map* M;


void setup()
{
    fprintf(stderr, "generating %zu keys ... ", n);
    xs = malloc(n * sizeof(char*));
    size_t i;
    size_t m;
    for (i = 0; i < n; ++i) {
        m = m_low + rand() % (m_high - m_low);
        xs[i] = malloc(m + 1);
        randstr(xs[i], m);
    }

    T = ahtable_create();
    M = str_map_create();
    fprintf(stderr, "done.\n");
}


void teardown()
{
    ahtable_free(T);
    str_map_destroy(M);

    size_t i;
    for (i = 0; i < n; ++i) {
        free(xs[i]);
    }
    free(xs);
}


bool test_ahtable_insert()
{
    fprintf(stderr, "inserting %zu keys ... \n", k);
    bool passed = true;
    size_t i, j;
    value_t* u;
    value_t  v;

    for (j = 0; j < k; ++j) {
        i = rand() % n;


        v = 1 + str_map_get(M, xs[i], strlen(xs[i]));
        str_map_set(M, xs[i], strlen(xs[i]), v);


        u = ahtable_get(T, xs[i], strlen(xs[i]));
        *u += 1;


        if (*u != v) {
            fprintf(stderr, "[error] tally mismatch (reported: %lu, correct: %lu)\n",
                            *u, v);
            passed = false;
        }
    }

    fprintf(stderr, "sizeof: %zu\n", ahtable_sizeof(T));

    /* delete some keys */
    for (j = 0; i < k/100; ++j) {
        i = rand() % n;
        ahtable_del(T, xs[i], strlen(xs[i]));
        str_map_del(M, xs[i], strlen(xs[i]));
        u = ahtable_tryget(T, xs[i], strlen(xs[i]));
        if (u) {
            fprintf(stderr, "[error] deleted node found in ahtable\n");
            passed = false;
        }
    }

    fprintf(stderr, "done.\n");
    return passed;
}


bool test_ahtable_iteration()
{
    fprintf(stderr, "iterating through %zu keys ... \n", k);

    ahtable_iter_t* i = ahtable_iter_begin(T, false);
    bool passed = true;
    size_t count = 0;
    value_t* u;
    value_t  v;

    size_t len;
    const char* key;

    while (!ahtable_iter_finished(i)) {
        ++count;

        key = ahtable_iter_key(i, &len);
        u   = ahtable_iter_val(i);
        v   = str_map_get(M, key, len);

        if (*u != v) {
            if (v == 0) {
                fprintf(stderr, "[error] incorrect iteration (%lu, %lu)\n", *u, v);
                passed = false;
            }
            else {
                fprintf(stderr, "[error] incorrect iteration tally (%lu, %lu)\n", *u, v);
                passed = false;
            }
        }

        // this way we will see an error if the same key is iterated through
        // twice
        str_map_set(M, key, len, 0);

        ahtable_iter_next(i);
    }

    if (count != M->m) {
        fprintf(stderr, "[error] iterated through %zu element, expected %zu\n",
                count, M->m);
        passed = false;
    }

    ahtable_iter_free(i);

    fprintf(stderr, "done.\n");
    return passed;
}


int cmpkey(const char* a, size_t ka, const char* b, size_t kb)
{
    int c = memcmp(a, b, ka < kb ? ka : kb);
    return c == 0 ? (int) ka - (int) kb : c;
}


bool test_ahtable_sorted_iteration()
{
    fprintf(stderr, "iterating in order through %zu keys ... \n", k);

    ahtable_iter_t* i = ahtable_iter_begin(T, true);
    bool passed = true;
    size_t count = 0;
    value_t* u;
    value_t  v;

    char* prev_key = malloc(m_high + 1);
    size_t prev_len = 0;

    const char *key = NULL;
    size_t len = 0;

    while (!ahtable_iter_finished(i)) {
        memcpy(prev_key, key, len);
        prev_len = len;
        ++count;

        key = ahtable_iter_key(i, &len);
        if (prev_key != NULL && cmpkey(prev_key, prev_len, key, len) > 0) {
            fprintf(stderr, "[error] iteration is not correctly ordered.\n");
            passed = false;
        }

        u  = ahtable_iter_val(i);
        v  = str_map_get(M, key, len);

        if (*u != v) {
            if (v == 0) {
                fprintf(stderr, "[error] incorrect iteration (%lu, %lu)\n", *u, v);
                passed = false;
            }
            else {
                fprintf(stderr, "[error] incorrect iteration tally (%lu, %lu)\n", *u, v);
                passed = false;
            }
        }

        // this way we will see an error if the same key is iterated through
        // twice
        str_map_set(M, key, len, 0);

        ahtable_iter_next(i);
    }

    ahtable_iter_free(i);
    free(prev_key);

    fprintf(stderr, "done.\n");
    return passed;
}

bool test_ahtable_save_load()
{
    fprintf(stderr, "saving ahtable ... \n");

    bool passed = true;
    FILE* fd_w = fopen("test.aht", "w");
    ahtable_save(T, fd_w);
    fclose(fd_w);

    fprintf(stderr, "loading ahtable ... \n");

    FILE* fd_r = fopen("test.aht", "r");
    ahtable_t* U = ahtable_load(fd_r);
    fclose(fd_r);

    fprintf(stderr, "comparing ahtable ... \n");

    ahtable_iter_t* i = ahtable_iter_begin(T, false);
    ahtable_iter_t* j = ahtable_iter_begin(U, false);
    const char *k1 = NULL;
    const char *k2 = NULL;
    value_t* v1;
    value_t* v2;
    size_t len1 = 0;
    size_t len2 = 0;
    while (!ahtable_iter_finished(i) && !ahtable_iter_finished(j)) {
        k1 = ahtable_iter_key(i, &len1);
        v1 = ahtable_iter_val(i);

        k2 = ahtable_iter_key(j, &len2);
        v2 = ahtable_iter_val(j);

        if (len1 != len2) {
            fprintf(stderr, "[error] key lengths don't match (%lu, %lu)\n", len1, len2);
            passed = false;
        } else if (strncmp(k1, k2, len1) != 0) {
            fprintf(stderr, "[error] key strings don't match (%s, %s)\n", k1, k2);
            passed = false;
        }

        if (*v1 != *v2) {
            fprintf(stderr, "[error] values don't match (%lu, %lu)\n", *v1, *v2);
            passed = false;
        }

        ahtable_iter_next(i);
        ahtable_iter_next(j);
    }
    ahtable_iter_free(i);
    ahtable_iter_free(j);
    return passed;
}


int main()
{
    bool passed = true;

    setup();
    passed &= test_ahtable_insert();
    passed &= test_ahtable_iteration();
    teardown();

    setup();
    passed &= test_ahtable_insert();
    passed &= test_ahtable_sorted_iteration();
    teardown();

    setup();
    passed &= test_ahtable_insert();
    passed &= test_ahtable_save_load();
    teardown();

    if (passed) return 0;
    return 1;
}
