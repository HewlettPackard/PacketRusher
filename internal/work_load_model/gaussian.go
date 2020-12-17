package work_load_model

/*
#cgo CFLAGS: -I /usr/local/include
#cgo LDFLAGS: -L /usr/local/lib -lgsl -lgslcblas -lm
#include<gsl/gsl_rng.h>
#include<gsl/gsl_randist.h>

unsigned int randomNumberFromGaussian(double mu, long seed){
	const gsl_rng_type* T;
    gsl_rng* r;
    unsigned int dist;

	gsl_rng_env_setup();
	T = gsl_rng_default;
	r = gsl_rng_alloc(T);

    gsl_rng_set(r,seed);
    dist = gsl_ran_gaussian (r, mu);

	gsl_rng_free (r);

	return dist;
}
*/
import "C"

// calculate some random numbers from the Gaussian distribution
func gaussianlDistribution(sigma float64, length int, const_seed int) (distGaussian []uint) {
	distGaussian = make([]uint, length)

	for i := 1; i <= length; i++ {
		seed := const_seed + i
		aux := C.randomNumberFromGaussian(C.double(sigma), C.long(seed))
		distGaussian[i-1] = uint(aux)
	}
	return
}
