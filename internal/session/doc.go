// Package session keeps one applicant's answers between requests.
//
// [Store] is the persistence seam (the example uses an in-memory map). Sessions
// carry a CSRF token, cookie consent choice, and journey data. Swap the store
// implementation when a real service needs Redis or SQL — keep the interface.
package session
