#pragma once

// extern "C" {

int m115_edinit();

int m115_xorinit();

int m115_encode(unsigned char const*, unsigned int, unsigned char*, unsigned int*, unsigned char*, unsigned char*);

int m115_decode(unsigned char const*, unsigned int, unsigned char*, unsigned int*, unsigned char*, unsigned char*);

// }
